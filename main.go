package main

import (
	"bytes"
	"context"
	"database/sql"
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
	database, err := db.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	address := fmt.Sprintf("%s:%d", cCtx.String("host"), cCtx.Uint64("port"))

	mux := http.NewServeMux()
	ah := audioHandler()
	fh := feedHandler(database)
	gh := genFeedHandler(database, cCtx)
	mux.Handle("/audio/", ah)
	mux.Handle("/feed/", fh)
	mux.Handle("/gen/", gh)
	err = http.ListenAndServe(address, mux)
	return err
}

func feedHandler(database *sql.DB) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		feedId := strings.TrimPrefix(r.URL.Path, "/feed/")
		xml, err := db.GetFeed(context.Background(), database, &feedId)
		if err == sql.ErrNoRows {
			http.Error(w, "Feed not found, please generate it.", http.StatusNotFound)
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(*xml))
		}
	}
	return http.HandlerFunc(fn)
}

func genFeedHandler(database *sql.DB, cCtx *cli.Context) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ytPathPart := strings.TrimPrefix(r.URL.Path, "/gen/")
		feedUrl := genFeed(ytPathPart, database, cCtx)
		w.Write([]byte(feedUrl))
	}
	return http.HandlerFunc(fn)
}

func genFeed(ytPathPart string, database *sql.DB, cCtx *cli.Context) string {
	base_url := cCtx.String("base-url")

	youtubeUrl := fmt.Sprintf("https://www.youtube.com/%s", ytPathPart)
	cmd := exec.Command("yt-dlp", "-J", "--playlist-items=:20", youtubeUrl)
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
	podcastFeed := podcast.New(c.Title, c.Url, c.Description, &now, &now)
	podcastFeed.AddAuthor(c.Author, "")
	podcastFeed.AddImage(getFeedImage(&c))
	podcastFeed.AddSummary(c.Description)
	podcastFeed.IExplicit = "no"
	podcastFeed.IBlock = "Yes"
	podcastFeed.Generator = "vpod"

	for _, v := range c.Videos {
		var enclosureUrl string
		var enclosureLengthBytes int64
		acceptable_file_found := false
		for _, f := range v.Formats {
			audio_only := f.Resolution == "audio only"
			correct_ext := f.AudioExt == "m4a"
			no_drm := !f.Drm
			no_dynamic_range_compression := !strings.Contains(f.Id, "drc")

			if audio_only && correct_ext && no_drm && no_dynamic_range_compression {
				acceptable_file_found = true
				enclosureUrl = fmt.Sprintf("%s/audio/%s/%s", base_url, v.Id, f.Id)
				enclosureLengthBytes = f.Filesize
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
		item.AddEnclosure(enclosureUrl, podcast.M4A, enclosureLengthBytes)

		if _, err := podcastFeed.AddItem(item); err != nil {
			log.Fatal(err)
		}
	}

	feedXML := new(bytes.Buffer)
	podcastFeed.Encode(feedXML)
	feedXMLStr := feedXML.String()
	db.CreateFeed(context.Background(), database, &c.Id, &c.Title, &c.Description, &c.Url, &feedXMLStr)

	finalFeedUrl := fmt.Sprintf("%s/feed/%s", base_url, c.Id)
	return finalFeedUrl
}

func getFeedImage(c *YouTubeChannel) string {
	for _, logo := range c.Logos {
		if logo.Preference == 1 {
			return logo.Url
		}
	}
	return "https://upload.wikimedia.org/wikipedia/commons/thumb/5/59/Minecraft_missing_texture_block.svg/1024px-Minecraft_missing_texture_block.svg.png"

}
