package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/eduncan911/podcast"
)

const placeholder_image = "https://upload.wikimedia.org/wikipedia/commons/thumb/5/59/Minecraft_missing_texture_block.svg/1024px-Minecraft_missing_texture_block.svg.png"

func getChannel(ytURL url.URL, numItems uint64) (*YouTubeChannel, error) {
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

	var c YouTubeChannel
	err = json.Unmarshal(outb.Bytes(), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func channelToPodcast(
	c YouTubeChannel,
	pubDate *time.Time,
	baseURL *url.URL,
) (*Podcast, error) {
	now := time.Now()
	p := podcast.New(
		strings.Replace(c.Title, " - Videos", "", -1),
		c.Url, c.Description, pubDate, &now, // TODO: fix the distinction in the db between update_time and build time here
	)
	imageLink := getFeedImage(&c)
	p.AddAuthor(c.Author, "")
	p.AddImage(imageLink.String())
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
				enclosureUrl = fmt.Sprintf("%s/audio/%s/%s", baseURL, v.Id, f.Id)
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

	return &Podcast{
		Podcast: &p,
		Id:      c.Id,
	}, nil
}

func fetchPodcast(ytURL url.URL, numItems uint64, ctx context.Context) (*Podcast, error) {
	u, ok := ctx.Value("baseURL").(string)
	if !ok {
		return nil, errors.New("could not get baseURL from ctx")
	}
	baseURL, err := url.Parse(u)
	if err != nil {
		return nil, errors.New("could not parse base-url as url.URL")
	}

	c, err := getChannel(ytURL, numItems)
	if err != nil {
		return nil, err
	}

	// pubDate, _ := time.Parse(time.RFC1123Z, "Mon, 02 Jan 2006 15:04:05 -0700")
	pubDate := new(time.Time)

	return channelToPodcast(*c, pubDate, baseURL) // TODO: decide what to do about PubDate
}

func getFeedImage(c *YouTubeChannel) url.URL {
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
	return *u
}
