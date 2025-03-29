package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eduncan911/podcast"
	"log"
	"os/exec"
	"time"
)

func main() {
	cmd := exec.Command("yt-dlp", "-J", "https://www.youtube.com/@Monoanalysis")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	var c YouTubeChannel
	err = json.Unmarshal(out.Bytes(), &c)
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()
	podcast_feed := podcast.New(c.Title, c.Url, c.Description, &now, &now)

	for i := 0; i < len(c.Playlists[0].Videos); i++ {
		v := c.Playlists[0].Videos[i]

		i := podcast.Item{
			Title:       v.Title,
			Description: v.Description,
			Link:        v.Url,
		}
		d := v.ReleaseTimestamp.Time
		i.AddPubDate(&d)

		if _, err := podcast_feed.AddItem(i); err != nil {
			log.Fatal(err)
		}

		fmt.Println(i.Title)
		fmt.Println(i.Description)
		fmt.Println(i.Link)
		fmt.Println(i.PubDateFormatted)
	}
}
