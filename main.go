package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/eduncan911/podcast"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"vpod/db"
)

func main() {
	app := &cli.App{
		Name:  "vpod",
		Usage: "beware the pipeline",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "base-url",
				Usage: "The base url for the podcast",
			},
		},
		Action: func(cCtx *cli.Context) error {
			genFeed(cCtx.String("base-url"))
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func genFeed(base_url string) {
	database, err := db.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	cmd := exec.Command("yt-dlp", "-J", "https://www.youtube.com/@Monoanalysis")
	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	var c YouTubeChannel
	err = json.Unmarshal(out.Bytes(), &c)
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()
	db.CreateFeed(context.Background(), database, &c.Title, &c.Id, &c.Description, &c.Url)
	podcast_feed := podcast.New(c.Title, c.Url, c.Description, &now, &now)
	podcast_feed.AddSummary(c.Description)
	podcast_feed.AddAuthor(c.Author, "")
	podcast_feed.IExplicit = "no"
	podcast_feed.IBlock = "Yes"
	podcast_feed.Generator = "vpod"

	for i := 0; i < len(c.Playlists[0].Videos); i++ {
		v := c.Playlists[0].Videos[i]

		acceptable_file_found := false
		var enclosureUrl string
		var enclosureFilesize int64
		for i := 0; i < len(v.Formats); i++ {
			f := v.Formats[i]
			audio_only := f.Resolution == "audio only"
			correct_ext := f.AudioExt == "m4a"
			no_drm := !f.Drm
			no_dynamic_range_compression := !strings.Contains(f.Id, "drc")

			if audio_only && correct_ext && no_drm && no_dynamic_range_compression {
				acceptable_file_found = true
				enclosureUrl = fmt.Sprintf("%s/%s/%s", base_url, v.Id, f.Id)
				enclosureFilesize = f.Filesize
				break
			}
		}
		if !acceptable_file_found {
			fmt.Println("No acceptable file found, moving on.")
			continue
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
		item.AddEnclosure(enclosureUrl, podcast.M4A, enclosureFilesize)

		if _, err := podcast_feed.AddItem(item); err != nil {
			log.Fatal(err)
		}
	}

	podcast_feed.Encode(os.Stdout)
}
