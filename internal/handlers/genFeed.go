package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"vpod/internal/data"
	"vpod/internal/podcast"
	"vpod/internal/views"
	"vpod/internal/youtube"

	"github.com/urfave/cli/v2"
)

// allowedYouTubeHosts is the set of hostnames accepted for channel URLs.
var allowedYouTubeHosts = map[string]bool{
	"youtube.com":     true,
	"www.youtube.com": true,
	"m.youtube.com":   true,
	"youtu.be":        true,
	"www.youtu.be":    true,
}

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
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if ytURL.Scheme != "https" && ytURL.Scheme != "http" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", ytURL.Scheme)
	}
	if !allowedYouTubeHosts[ytURL.Hostname()] {
		return nil, fmt.Errorf("URL host %q is not an allowed YouTube domain", ytURL.Hostname())
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
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		baseURL, err := url.Parse(cCtx.String("base-url"))
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Could not parse baseURL from context")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		channelURL := r.FormValue("channelURL")
		if channelURL == "" {
			logger.Error("channelURL is blank")
			http.Error(w, "channelURL cannot be blank", http.StatusBadRequest)
			return
		}

		p, err := gen(ctx, channelURL, baseURL, logger, queries)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when generating feed.")
			http.Error(w, "Failed to generate feed", http.StatusInternalServerError)
			return
		}
		logger.Debug("Feed successfully generated")

		u := baseURL.JoinPath("feed", p.Id)
		feedURL := u.String()
		data := FeedPageData{
			Image:          p.Image.URL,
			Title:          p.Title,
			URL:            feedURL,
			URLNoScheme:    stripScheme(feedURL),
			URLPathEscaped: url.PathEscape(feedURL),
		}
		// Path is relative to where command runs
		tmpl := template.Must(template.ParseFS(views.ViewFS, "podcastSuccess.html"))
		tmpl.Execute(w, data)
	}
	return http.HandlerFunc(fn)
}

// FeedPageData contains data for rendering the success page after creating a feed.
type FeedPageData struct {
	Image          string
	Title          string
	URL            string
	URLNoScheme    string // URL without scheme for podcast:// style links
	URLPathEscaped string
}
