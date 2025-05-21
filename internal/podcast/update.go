package podcast

import (
	"context"
	"errors"
	"vpod/internal/data"

	"github.com/eduncan911/podcast"
)

func (p Podcast) AppendOldEps(ctx context.Context) (*Podcast, error) {
	queries, ok := ctx.Value("queries").(*data.Queries)
	if !ok {
		return nil, errors.New("could not get queries from ctx")
	}

	numItems := len(p.Items)
	if numItems < 1 {
		return nil, errors.New("detected 0 episodes in feed")
	}

	latestEp := p.Items[numItems-1]

	oldEps, err := queries.GetOlderEpisodesForFeed(ctx, data.GetOlderEpisodesForFeedParams{
		FeedID: p.Id,
		ID:     []byte(latestEp.GUID),
	})
	if err != nil {
		return nil, err
	}

	for _, ep := range oldEps {
		item := podcast.Item{
			Title:       ep.Title,
			Description: ep.Description.String,
			Link:        ep.VideoUrl.String,
		}
		item.AddPubDate(&ep.ReleasedAt.Time)
		item.AddDuration(ep.Duration.Int64)
		item.AddImage(ep.Thumbnail.String)
		item.AddEnclosure(ep.AudioUrl, podcast.M4A, ep.AudioLengthBytes)

		if _, err := p.AddItem(item); err != nil {
			return nil, err
		}
	}

	return &p, nil
}
