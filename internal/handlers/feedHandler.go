package handlers

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"vpod/internal/data"
)

func FeedLegacy(queries *data.Queries) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value("logger").(*slog.Logger)

		feedId := strings.TrimPrefix(r.URL.Path, "/feed/")
		logger = logger.With(slog.String("feed_id", feedId))

		logger.Info("Getting feed from DB")
		xml, err := queries.GetFeedXML(ctx, []byte(feedId))

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
