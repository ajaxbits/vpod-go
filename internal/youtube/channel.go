package youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

type UnixTime struct {
	time.Time
}

func (u *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	u.Time = time.Unix(timestamp, 0)
	return nil
}

type Channel struct {
	Author      string `json:"uploader"`
	Description string
	Id          string        `json:"channel_id"`
	Logos       []ChannelLogo `json:"thumbnails"`
	Videos      []Video       `json:"entries"`
	Title       string
	URL         url.URL `json:"channel_url"`
}

func (c *Channel) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Author      string `json:"uploader"`
		Description string
		Entries     json.RawMessage `json:"entries"`
		Id          string          `json:"channel_id"`
		Logos       []ChannelLogo   `json:"thumbnails"`
		Title       string
		URL         string `json:"channel_url"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	u, err := url.Parse(tmp.URL)
	if err != nil {
		return err
	}
	c.URL = *u

	var firstEntry []map[string]interface{}
	if err := json.Unmarshal(tmp.Entries, &firstEntry); err != nil {
		return err
	}
	eType, typeKeyExists := firstEntry[0]["_type"].(string)
	if typeKeyExists && eType == "playlist" {
		var playlists []Playlist
		if err := json.Unmarshal(tmp.Entries, &playlists); err != nil {
			return err
		}

		var videosPlaylist Playlist
		for _, playlist := range playlists {
			if strings.Contains(strings.ToLower(playlist.Title), "videos") {
				videosPlaylist = playlist
			}
		}
		c.Videos = videosPlaylist.Videos
	} else {
		var videos []Video
		if err := json.Unmarshal(tmp.Entries, &videos); err != nil {
			return err
		}
		c.Videos = videos
	}

	c.Author = tmp.Author
	c.Description = tmp.Description
	c.Id = tmp.Id
	c.Logos = tmp.Logos
	c.Title = tmp.Title
	return nil
}

type ChannelLogo struct {
	Id         string
	Preference int `json:"preference,omitempty"`
	Url        string
}

func (c *Channel) GetLogo() *url.URL {
	const placeholder_image = "https://upload.wikimedia.org/wikipedia/commons/thumb/5/59/Minecraft_missing_texture_block.svg/1024px-Minecraft_missing_texture_block.svg.png"

	urlStr := placeholder_image
	for _, logo := range c.Logos {
		if logo.Preference == 1 {
			urlStr = logo.Url
		}
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	return u
}

type fetchChannelOptions struct {
	numItems *uint64
}

type FetchChannelOption func(options *fetchChannelOptions) error

func WithNItems(n uint64) FetchChannelOption {
	return func(options *fetchChannelOptions) error {
		if n == 0 {
			return errors.New("must fetch at least one video")
		}
		options.numItems = &n
		return nil
	}
}

func FetchChannel(ytURL url.URL, opts ...FetchChannelOption) (*Channel, error) {
	var options fetchChannelOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	var numItems uint64
	if options.numItems == nil {
		numItems = uint64(5)
	} else {
		numItems = *options.numItems
	}

	cmd := exec.Command(
		"yt-dlp",
		"-J",
		fmt.Sprintf("--playlist-items=0:%d", numItems),
		ytURL.String(),
	)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return nil, errors.New(errb.String())
		} else {
			return nil, err
		}
	}

	var c Channel
	err = json.Unmarshal(outb.Bytes(), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
