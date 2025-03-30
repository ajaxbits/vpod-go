package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/eduncan911/podcast"
	"log"
	"os/exec"
	"time"

	"vpod/db"
)

func main() {
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
	podcast_feed.IExplicit = "no"

	for i := 0; i < len(c.Playlists[0].Videos); i++ {
		v := c.Playlists[0].Videos[i]

		item := podcast.Item{
			Title:       v.Title,
			Description: v.Description,
			Link:        v.Url,
		}
		d := v.ReleaseTimestamp.Time
		item.AddPubDate(&d)

		acceptable_file_found := false
		for i := 0; i < len(v.Formats); i++ {
			f := v.Formats[i]
			audio_only := f.Resolution == "audio only"
			correct_ext := f.AudioExt == "m4a"
			no_drm := !f.Drm

			if audio_only && correct_ext && no_drm {
				acceptable_file_found = true
				item.AddEnclosure(f.Url, podcast.M4A, f.Filesize) // TODO: set proper url here
				break
			}
		}

		if !acceptable_file_found {
			fmt.Println("No acceptable file found, moving on.")
			continue
		}

		if _, err := podcast_feed.AddItem(item); err != nil {
			log.Fatal(err)
		}

		fmt.Println(item.Title)
		fmt.Println(item.Enclosure)
	}
}
