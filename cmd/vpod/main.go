package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"vpod/internal/podcast"
	"vpod/internal/youtube"

	"github.com/urfave/cli/v2"
)

type CliFlags struct {
	BaseUrl string
	Port    int64
}

func main() {
	app := &cli.App{
		Name:  "vpod",
		Usage: "beware the pipeline",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "base-url",
				Usage:   "The base url for the podcast",
				EnvVars: []string{"BASE_URL"},
			},
			&cli.StringFlag{
				Name:    "host",
				Usage:   "The addres to run the web server on",
				Value:   "0.0.0.0",
				EnvVars: []string{"HOST"},
			},
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "Log level for the program",
				Value:   "INFO",
				EnvVars: []string{"LOG_LEVEL"},
			},
			&cli.Uint64Flag{
				Name:  "port",
				Usage: "The port to run the web server on.",

				Value:   8080,
				EnvVars: []string{"PORT"},
				Action: func(ctx *cli.Context, v uint64) error {
					if v >= 65536 {
						return fmt.Errorf("Invalid port: %v. Must be in range[0-65535]", v)
					}
					return nil
				},
			},
		},
		Action: func(cCtx *cli.Context) error {
			err := serve(cCtx)
			return err
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(cCtx *cli.Context) error {
	env, err := NewEnv(
		cCtx.String("log-level"),
		cCtx.String("base-url"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer env.Cleanup()

	logger := env.logger
	logger.Debug("Env initalized")

	mux := http.NewServeMux()
	mux.Handle("/audio/", audioHandler())
	mux.Handle("/feed/", env.feedHandler())
	mux.Handle("/gen/", env.genFeedHandler(cCtx))

	address := fmt.Sprintf("%s:%d", cCtx.String("host"), cCtx.Uint64("port"))
	handler := loggingWrapper(mux, logger)
	srv := &http.Server{
		Addr:         address,
		ReadTimeout:  300 * time.Second, // for long audio returns
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  300 * time.Second,
		Handler:      handler,
	}
	logger.Info("starting server", slog.String("address", address))
	return srv.ListenAndServe()
}

func (e *Env) feedHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value("logger").(*slog.Logger)

		feedId := strings.TrimPrefix(r.URL.Path, "/feed/")
		logger = logger.With(slog.String("feed_id", feedId))

		logger.Info("Getting feed from DB")
		xml, err := e.queries.GetFeedXML(ctx, []byte(feedId))

		if err == sql.ErrNoRows {
			logger.Error("Feed not found in Database")
			http.Error(w, "Feed not found, please generate it.", http.StatusNotFound)
		} else if err != nil {
			logger.With(slog.String("err", fmt.Sprintf("%v", err))).Error("Something went wrong when fetching feed.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			logger.Debug("Feed found in DB")
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(xml))
		}
	}
	return http.HandlerFunc(fn)
}

func (e *Env) genFeedHandler(cCtx *cli.Context) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, "url", r.URL)
		logger := ctx.Value("logger").(*slog.Logger)

		baseURL, err := url.Parse(cCtx.String("base-url"))
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when generating feed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("generating feed")

		ytURL := url.URL{
			Scheme: "https",
			Host:   "youtube.com",
			Path:   strings.TrimPrefix(r.URL.Path, "/gen/"),
		}
		c, err := youtube.FetchChannel(&ytURL, youtube.WithNItems(20))
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when fetching feed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p, err := podcast.FromChannel(*c, *baseURL) // TODO: decide what to do about PubDate
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when generating feed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = upsertPodcast(e.queries, *p, ctx)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when inserting feed into db")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		feedURL := baseURL.JoinPath("feed", p.Id)
		logger.Debug("Feed successfully generated")
		w.Write([]byte(feedURL.String()))
	}
	return http.HandlerFunc(fn)
}

func (e *Env) updateHandler(baseURL string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		const defaultNewEpsToFetch = 5

		ctx := r.Context()
		ctx = context.WithValue(ctx, "baseURL", baseURL)
		ctx = context.WithValue(ctx, "queries", e.queries)
		logger := ctx.Value("logger").(*slog.Logger)

		baseURL, err := url.Parse(baseURL)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when generating feed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		feedId := strings.TrimPrefix(r.URL.Path, "/update/")
		logger = logger.With(slog.String("feed_id", feedId))

		ytURL := url.URL{
			Scheme: "https",
			Host:   "www.youtube.com",
			Path:   fmt.Sprintf("/channel/%s", feedId),
		}
		c, err := youtube.FetchChannel(&ytURL)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when fetching feed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p, err := podcast.FromChannel(*c, *baseURL) // TODO: decide what to do about PubDate
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when generating feed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p, err = p.AppendOldEps(ctx)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when adding old eposides to feed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = upsertPodcast(e.queries, *p, ctx)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when inserting feed into db")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return http.HandlerFunc(fn)
}
