package handlers

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"vpod/internal/data"

	"github.com/urfave/cli/v2"
)

const PAGESIZE = 10

func getFeedPage(q *data.Queries, ctx context.Context, pageNumber uint64) ([]data.Feed, error) {
	params := data.GetAllFeedsParams{
		Limit:   PAGESIZE,
		Column2: pageNumber,
	}
	return q.GetAllFeeds(ctx, params)
}

// TODO unit test babyyyyy
func getPage(u *url.URL) (uint64, error) {
	pageStr := u.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	page, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return page, nil
}

func list(
	ctx context.Context,
	baseURL *url.URL,
	logger *slog.Logger,
	queries *data.Queries,
	page uint64,
) (*[]FeedListEntry, error) {
	logger.Info("Getting Feeds")
	feeds, err := getFeedPage(queries, ctx, page)
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

		page, err := getPage(r.URL)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when getting all the feeds.")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		entries, err := list(ctx, baseURL, logger, queries, page)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when getting all the feeds.")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		data := FeedListData{
			Entries:  *entries,
			NextPage: page + 1,
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

type FeedListData struct {
	Entries  []FeedListEntry
	NextPage uint64
}

type FeedListEntry struct {
	ChannelURL  string
	Description string
	LastUpdated time.Time
	NumEps      uint64
	Title       string
	URL         string
}
