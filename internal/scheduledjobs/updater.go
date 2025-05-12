package scheduledjobs

import (
	"context"
	"log/slog"
	"net/url"
	"vpod/internal/data"
	"vpod/internal/podcast"
	"vpod/internal/youtube"
)

func update(
	ctx context.Context,
	feedID string,
	baseURL *url.URL,
	queries *data.Queries,
) error {
	ytURL := &url.URL{
		Scheme: "https",
		Host:   "www.youtube.com",
	}
	ytURL = ytURL.JoinPath("channel", feedID)
	c, err := youtube.FetchChannel(ytURL)
	if err != nil {
		return err
	}

	p, err := podcast.FromChannel(*c, *baseURL) // TODO: decide what to do about PubDate
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, "queries", queries) // TODO: smelly
	p, err = p.AppendOldEps(ctx)
	if err != nil {
		return err
	}

	err = podcast.UpsertPodcast(queries, *p, ctx)
	if err != nil {
		return err
	}
	return nil
}

func updateAll(
	ctx context.Context,
	logger *slog.Logger,
	baseURL *url.URL,
	queries *data.Queries,
) error {
	ids, err := queries.GetAllFeedIds(ctx)
	if err != nil {
		logger.Error(
			"could not get feeds from DB",
			slog.String("err", err.Error()),
		)
		return err
	}
	for _, idBytes := range ids {
		if idBytes == nil {
			logger.Warn("got a nil id from the DB")
			continue
		}
		id := string(idBytes)
		logger.Debug(
			"updating feed",
			slog.String("feed_id", id),
		)
		err = update(ctx, id, baseURL, queries)
		if err != nil {
			logger.Error(
				"could not update feed",
				slog.String("err", err.Error()),
			)
			continue // TODO: handle this case
		}
		logger.Info(
			"updated feed",
			slog.String("feed_id", id),
		)
	}
	return nil
}
