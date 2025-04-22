package podcast

import (
	"errors"
	"net/url"
	"time"

	"github.com/eduncan911/podcast"
)

type Podcast struct {
	*podcast.Podcast
	Id string
}

func New(
	id string,
	title string,
	link url.URL,
	description string,
	pubDate *time.Time,
	lastBuildDate *time.Time,
) (*Podcast, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	p := podcast.New(
		title,
		link.String(),
		description,
		pubDate,
		lastBuildDate,
	)

	return &Podcast{
		Id:      id,
		Podcast: &p,
	}, nil
}
