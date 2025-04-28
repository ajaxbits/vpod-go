package podcast

import (
	"errors"
	"net/url"

	"strings"
	"time"
	"vpod/internal/youtube"

	"github.com/eduncan911/podcast"
)

type options struct {
	pubDate       *time.Time
	lastBuildDate *time.Time
}

type Option func(options *options) error

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
	description string,
	opts ...Option,
) (*Podcast, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}
	if link.String() == "" {
		return nil, errors.New("link cannot be empty")
	}

	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	var (
		lastBuildDate time.Time
		pubDate       time.Time
	)
	if options.lastBuildDate == nil {
		lastBuildDate = time.Now().UTC()
		// TODO: fix the distinction in the db between update_time and build time here
	} else {
		lastBuildDate = *options.lastBuildDate
	}
	if options.pubDate == nil {
		pubDate = time.Now().UTC()
	} else {
		pubDate = *options.pubDate
	}

	p := podcast.New(
		title,
		link.String(),
		description,
		&pubDate,
		&lastBuildDate,
	)

	return &Podcast{
		Id:      id,
		Podcast: &p,
	}, nil
}

func FromChannel(c youtube.Channel, baseURL url.URL, opts ...Option) (*Podcast, error) {
	p, err := New(
		c.Id,
		strings.Replace(c.Title, " - Videos", "", -1),
		c.URL,
		c.Description,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	p.AddAuthor(c.Author, "no_email_provided") // No kidding, we must add an email of len > 0...
	p.AddImage(c.GetLogo().String())
	p.AddSummary(c.Description)
	p.IExplicit = "no"
	p.IBlock = "Yes"
	p.Generator = "vpod"

	for _, v := range c.Videos {
		var enclosureUrl string
		var enclosureLengthBytes int64
		acceptable_file_found := false
		for _, f := range v.Formats {
			is_english := strings.Split(f.Language, "-")[0] == "en"
			audio_only := f.Resolution == "audio only"
			correct_ext := f.AudioExt == "m4a"
			no_drm := !f.Drm
			no_dynamic_range_compression := !strings.Contains(f.Id, "drc")

			if is_english && audio_only && correct_ext && no_drm && no_dynamic_range_compression {
				acceptable_file_found = true
				enclosureUrl = baseURL.JoinPath("audio", v.Id, f.Id).String()
				enclosureLengthBytes = f.Filesize
				break
			}
		}
		if !acceptable_file_found {
			// TODO: figure out how to handle these situations
			continue
		}

		if v.Title == "" {
			v.Title = "untitled"
		}
		if v.Description == "" {
			v.Description = "no description provided"
		}

		item := podcast.Item{
			Title:       v.Title,
			Description: v.Description,
			Link:        v.Url,
		}
		d := v.ReleaseTimestamp.Time
		item.AddPubDate(&d)
		item.AddDuration(v.Duration)
		item.AddImage(v.Thumbnail)
		item.AddEnclosure(enclosureUrl, podcast.M4A, enclosureLengthBytes)

		if _, err := p.AddItem(item); err != nil {
			return nil, err
		}
	}

	return p, nil
}
