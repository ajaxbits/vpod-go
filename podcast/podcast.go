package podcast

import (
	"errors"
	"net/url"
	"time"

	"github.com/eduncan911/podcast"
)

type options struct {
	description   *string
	pubDate       *time.Time
	lastBuildDate *time.Time
}

type Option func(options *options) error

func WithDescription(d string) Option {
	return func(options *options) error {
		options.description = &d
		return nil
	}
}

func WithPubDate(p time.Time) Option {
	return func(options *options) error {
		options.pubDate = &p
		return nil
	}
}

type Podcast struct {
	*podcast.Podcast
	Id string
}

func New(
	id string,
	title string,
	link url.URL,
	opts ...Option,
) (*Podcast, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	var description string
	if options.description == nil {
		description = ""
	} else {
		description = *options.description
	}

	p := podcast.New(
		title,
		link.String(),
		description,
		options.pubDate,
		options.lastBuildDate,
	)

	return &Podcast{
		Id:      id,
		Podcast: &p,
	}, nil
}
