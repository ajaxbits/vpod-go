package handlers

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"time"
	"vpod/internal/data"

	"github.com/urfave/cli/v2"
)

func list(
	ctx context.Context,
	baseURL *url.URL,
	logger *slog.Logger,
	queries *data.Queries,
) (*[]FeedListEntry, error) {
	logger.Info("Getting Feeds")
	feeds, err := queries.GetAllFeeds(ctx, 100) // TODO: actually implement pagination
	if err != nil {
		return nil, err
	}

	feedListEntries := make([]FeedListEntry, 0, len(feeds))
	for _, feed := range feeds {
		feedListEntries = append(feedListEntries, FeedListEntry{
			ChannelURL:  feed.Link,
			Description: feed.Description.String,
			LastUpdated: feed.UpdatedAt.Time,
			NumEps:      0, // TODO
			Title:       feed.Title,
			URL:         baseURL.JoinPath("feed", string(feed.ID)).String(),
		})
	}
	return &feedListEntries, nil
}

func GetFeeds(cCtx *cli.Context, queries *data.Queries) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value("logger").(*slog.Logger)

		baseURL, err := url.Parse(cCtx.String("base-url"))
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Could not parse baseURL from context")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := list(ctx, baseURL, logger, queries)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when getting all the feeds.")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Path is relative to where command runs
		tmpl := template.Must(template.ParseFiles("internal/views/feedList.html"))
		err = tmpl.Execute(w, data)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Failed to execute feedList template")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	return http.HandlerFunc(fn)
}

type FeedListEntry struct {
	ChannelURL  string
	Description string
	LastUpdated time.Time
	NumEps      uint64
	Title       string
	URL         string
}
