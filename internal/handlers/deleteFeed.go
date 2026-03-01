package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"vpod/internal/data"
)

// DeleteFeed handles HTTP DELETE requests to remove a feed and its episodes.
// Returns 200 OK on success with an empty body to allow HTMX to remove the element.
func DeleteFeed(queries *data.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value("logger").(*slog.Logger)

		feedID := strings.TrimPrefix(r.URL.Path, "/ui/feeds/")
		if feedID == "" {
			http.Error(w, "feed ID required", http.StatusBadRequest)
			return
		}

		logger = logger.With(slog.String("feed_id", feedID))
		logger.Info("Deleting feed")

		// Delete episodes first due to foreign key constraint
		err := queries.DeleteEpisodesForFeed(ctx, feedID)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Failed to delete episodes for feed")
			http.Error(w, "failed to delete feed episodes", http.StatusInternalServerError)
			return
		}

		err = queries.DeleteFeed(ctx, []byte(feedID))
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Failed to delete feed")
			http.Error(w, "failed to delete feed", http.StatusInternalServerError)
			return
		}

		logger.Info("Feed deleted successfully")
		// Return empty response - HTMX will remove the element
		w.WriteHeader(http.StatusOK)
	}
}
