package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"vpod/internal/data"
	"vpod/internal/podcast"
	"vpod/internal/youtube"

	"github.com/urfave/cli/v2"
)

func GenFeedHandler(cCtx *cli.Context, queries *data.Queries) http.Handler {
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

		err = podcast.UpsertPodcast(queries, *p, ctx)
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
