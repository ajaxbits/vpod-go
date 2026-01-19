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
	"vpod/internal/views"

	"github.com/urfave/cli/v2"
)

func getFeedPage(q *data.Queries, ctx context.Context, pageSize uint, pageNumber uint64, sort string) ([]data.GetAllFeedsRow, error) {
	params := data.GetAllFeedsParams{
		Limit:     int64(pageSize),
		Page:      int64(pageNumber),
		SortOrder: sort,
	}
	return q.GetAllFeeds(ctx, params)
}

// TODO unit test babyyyyy
func getPage(u *url.URL) (uint64, string, error) {
	pageStr := u.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}

	sort := u.Query().Get("sort")
	if sort == "" {
		sort = "newest"
	}

	page, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil {
		return 0, "", err
	}

	return page, sort, nil
}

func getFeedListEntries(
	ctx context.Context,
	baseURL *url.URL,
	logger *slog.Logger,
	queries *data.Queries,
	pageSize uint,
	pageNum uint64,
	sort string,
) (*[]FeedListEntry, uint64, error) {
	logger.Info("Getting Feeds")

	nextPage := uint64(0)
	rows, err := getFeedPage(queries, ctx, pageSize, pageNum, sort)
	if err != nil {
		return nil, nextPage, err
	}

	feedListEntries := make([]FeedListEntry, 0, len(rows))
	if len(rows) > 0 {
		for _, row := range rows {
			feedListEntries = append(feedListEntries, FeedListEntry{
				ChannelURL:  row.Link,
				Description: row.Description.String,
				LastUpdated: row.UpdatedAt.Time,
				NumEps:      0, // TODO
				Title:       row.Title,
				URL:         baseURL.JoinPath("feed", string(row.ID)).String(),
			})
		}
		if rows[0].HasMore {
			nextPage = pageNum + 1
		}
	}
	return &feedListEntries, nextPage, nil
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

		page, sort, err := getPage(r.URL)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when getting the page number from the url.")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		pageSize := uint(10)
		entries, nextPage, err := getFeedListEntries(ctx, baseURL, logger, queries, pageSize, page, sort)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when getting all the feeds.")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		data := FeedListData{
			Entries:  *entries,
			NextPage: nextPage,
			Sort:     sort,
		}
		// Path is relative to where command runs
		tmpl := template.Must(template.ParseFS(views.ViewFS, "feedList.html"))
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
	NextPage uint64 // 0 means no next page
	Sort     string
}

type FeedListEntry struct {
	ChannelURL  string
	Description string
	LastUpdated time.Time
	NumEps      uint64
	Title       string
	URL         string
}
