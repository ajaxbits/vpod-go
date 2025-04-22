package podcast

import (
	"context"
	"errors"
	"vpod/data"

	"github.com/eduncan911/podcast"
)

func (p Podcast) AppendOldEps(ctx context.Context) (*Podcast, error) {
	queries, ok := ctx.Value("queries").(*data.Queries)
	if !ok {
		return nil, errors.New("could not get queries from ctx")
	}

	latestEp := p.Items[len(p.Items)-1]

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
