package handlers

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"vpod/internal/data"
	"vpod/internal/podcast"
	"vpod/internal/youtube"

	"github.com/urfave/cli/v2"
)

func gen(
	ctx context.Context,
	channelURL string,
	baseURL *url.URL,
	logger *slog.Logger,
	queries *data.Queries,
) (*podcast.Podcast, error) {
	logger.Info("generating feed")
	ytURL, err := url.Parse(channelURL)
	if err != nil {
		return nil, err
	}

	c, err := youtube.FetchChannel(ytURL, youtube.WithNItems(20))
	if err != nil {
		return nil, err
	}

	p, err := podcast.FromChannel(*c, *baseURL) // TODO: decide what to do about PubDate
	if err != nil {
		return nil, err
	}

	err = podcast.UpsertPodcast(queries, *p, ctx)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func GenFeed(cCtx *cli.Context, queries *data.Queries) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value("logger").(*slog.Logger)

		err := r.ParseForm()
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Could not parse form data")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		baseURL, err := url.Parse(cCtx.String("base-url"))
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Could not parse baseURL from context")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		channelURL := r.FormValue("channelURL")
		if channelURL == "" {
			err = errors.New("channelURL cannot be blank")
			logger.With(slog.String("err", err.Error())).Error("channelURL is blank")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		p, err := gen(ctx, channelURL, baseURL, logger, queries)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when generating feed.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Debug("Feed successfully generated")

		u := baseURL.JoinPath("feed", p.Id)
		data := FeedPageData{
			Image:          p.Image.URL,
			Title:          p.Title,
			URL:            u.String(),
			URLPathEscaped: url.PathEscape(u.String()),
		}
		// Path is relative to where command runs
		tmpl := template.Must(template.ParseFiles("internal/views/podcastSuccess.html"))
		tmpl.Execute(w, data)
	}
	return http.HandlerFunc(fn)
}

type FeedPageData struct {
	Image          string
	Title          string
	URL            string
	URLPathEscaped string
}
