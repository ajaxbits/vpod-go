package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
	"vpod/internal/api"
	"vpod/internal/handlers"
	"vpod/internal/middleware"
	"vpod/internal/router"

	"github.com/urfave/cli/v2"
)

func panicHandler(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("PANIC", slog.Any("err", err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
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

	wantedUser := cCtx.String("user")
	var wantedPass string
	if cCtx.String("password-file") != "" {
		path := cCtx.String("password-file")
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		wantedPass = strings.TrimSpace(string(contents))
	} else {
		wantedPass = cCtx.String("password")
	}

	r := router.New()
	r.Use(middleware.LogRequest(logger))
	r.Use(panicHandler(logger))

	r.HandleFunc("GET /audio/", handlers.Audio())
	r.HandleFunc("GET /feed/", handlers.Feed(env.queries))

	r.Group("/api", api.Routes)
	r.Group("/ui", func(r *router.Router) {
		if !cCtx.Bool("no-auth") {
			r.Use(middleware.NewBasicAuth(&wantedUser, &wantedPass))
		}

		// The trailing slash is important here
		// TODO: revisit after embeddings
		r.Handle("GET /static/", handlers.Static())

		r.HandleFunc("GET /", handlers.Index())
		r.HandleFunc("GET /feeds", handlers.GetFeeds(cCtx, env.queries))
		r.HandleFunc("POST /gen", handlers.GenFeed(cCtx, env.queries))
	})

	address := fmt.Sprintf("%s:%d", cCtx.String("host"), cCtx.Uint64("port"))
	srv := &http.Server{
		Addr:         address,
		ReadTimeout:  300 * time.Second, // for long audio returns
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  300 * time.Second,
		Handler:      r,
	}
	logger.Info("starting server", slog.String("address", address))
	return srv.ListenAndServe()
}
