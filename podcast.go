package main

import (
	"errors"

	"github.com/eduncan911/podcast"
)

type Podcast struct {
	*podcast.Podcast
	Id string
}

func NewPodcast(id string, p *podcast.Podcast) (*Podcast, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	if p == nil {
		return nil, errors.New("podcast cannot be nil")
	}
	return &Podcast{
		Id:      id,
		Podcast: p,
	}, nil
}
