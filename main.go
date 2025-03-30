package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/eduncan911/podcast"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"vpod/db"
)

type CliFlags struct {
	BaseUrl string
	Port    int64
}

func main() {
	app := &cli.App{
		Name:  "vpod",
		Usage: "beware the pipeline",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "base-url",
				Usage: "The base url for the podcast",
			},
			&cli.StringFlag{
				Name:  "host",
				Usage: "The addres to run the web server on",
				Value: "0.0.0.0",
			},
			&cli.Uint64Flag{
				Name:  "port",
				Usage: "The port to run the web server on.",
				Value: 8080,
			},
		},
		Action: func(cCtx *cli.Context) error {
			err := serve(cCtx)
			return err
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(cCtx *cli.Context) error {
	address := fmt.Sprintf("%s:%d", cCtx.String("host"), cCtx.Uint64("port"))

	mux := http.NewServeMux()
	ah := audioHandler()
	fh := feedHandler(cCtx)
	mux.Handle("/audio/", ah)
	mux.Handle("/feed/", fh)
	err := http.ListenAndServe(address, mux)
	return err
}

func feedHandler(cCtx *cli.Context) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ytPathPart := strings.TrimPrefix(r.URL.Path, "/feed/")
		feed := genFeed(ytPathPart, cCtx)

		w.Header().Set("Content-Type", "application/xml")
		err := feed.Encode(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	return http.HandlerFunc(fn)
}

func genFeed(ytPathPart string, cCtx *cli.Context) podcast.Podcast {
	database, err := db.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	base_url := cCtx.String("base-url")

	youtubeUrl := fmt.Sprintf("https://www.youtube.com/%s", ytPathPart)
	cmd := exec.Command("yt-dlp", "-J", youtubeUrl)
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
	podcastFeed := podcast.New(c.Title, c.Url, c.Description, &now, &now)
	podcastFeed.AddSummary(c.Description)
	podcastFeed.AddAuthor(c.Author, "")
	podcastFeed.IExplicit = "no"
	podcastFeed.IBlock = "Yes"
	podcastFeed.Generator = "vpod"

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
				enclosureUrl = fmt.Sprintf("%s/audio/%s/%s", base_url, v.Id, f.Id)
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

		if _, err := podcastFeed.AddItem(item); err != nil {
			log.Fatal(err)
		}
	}

	return podcastFeed
}
