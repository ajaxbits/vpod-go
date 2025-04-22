package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"vpod/internal/data"
	"vpod/internal/podcast"
)

func upsertPodcast(queries *data.Queries, p podcast.Podcast, ctx context.Context) error {
	pubDate, err := time.Parse(time.RFC1123Z, p.PubDate)
	if err != nil {
		return errors.New("could not parse podcast PubDate as RFC1123Z")
	}

	for _, i := range p.Items {
		err = upsertEpisode(i, &p.Id, queries, ctx)
		if err != nil {
			return err
		}
	}

	xml := new(bytes.Buffer)
	p.Encode(xml)

	return queries.UpsertFeed(ctx, data.UpsertFeedParams{
		ID: []byte(p.Id),
		CreatedAt: sql.NullTime{
			Time:  pubDate,
			Valid: true,
		},
		Description: sql.NullString{
			String: p.Description,
			Valid:  true,
		},
		Title: p.Title,
		Link:  p.Link,
		Xml:   xml.String(),
	})
}

func upsertEpisode(
	ep *podcast.Item,
	feedID *string,
	queries *data.Queries,
	ctx context.Context,
) error {
	duration, err := durationStrToInt(ep.IDuration)
	if err != nil {
		return err
	}

	err = queries.UpsertEpisode(ctx, data.UpsertEpisodeParams{
		ID:               []byte(ep.GUID), // TODO: make sure this is set to videoid
		AudioUrl:         ep.Enclosure.URL,
		AudioLengthBytes: ep.Enclosure.Length,
		Description:      sql.NullString{String: ep.Description, Valid: true},
		Duration: sql.NullInt64{
			Int64: duration,
			Valid: true,
		},
		FeedID: *feedID,
		ReleasedAt: sql.NullTime{
			Time:  *ep.PubDate,
			Valid: true,
		},
		Thumbnail: sql.NullString{
			String: ep.IImage.HREF,
			Valid:  true,
		},
		Title: ep.Title,
		VideoUrl: sql.NullString{
			String: ep.Link,
			Valid:  true,
		},
	})
	return err
}

func durationStrToInt(d string) (int64, error) {
	var h, m, s int64
	var err error

	parts := strings.Split(d, ":")
	switch len(parts) {
	case 3:
		// Format: H:MM:SS or HH:MM:SS
		h, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, err
		}
		m, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, err
		}
		s, err = strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return 0, err
		}
	case 2:
		// Format: M:SS or MM:SS
		m, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, err
		}
		s, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("invalid duration format: %s", d)
	}
	return h*3600 + m*60 + s, nil

}
