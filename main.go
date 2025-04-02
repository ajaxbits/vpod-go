package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eduncan911/podcast"
	"github.com/urfave/cli/v2"
	"log"
	"log/slog"
	"net/http"
	"net/url"
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
				Name:    "base-url",
				Usage:   "The base url for the podcast",
				EnvVars: []string{"BASE_URL"},
			},
			&cli.StringFlag{
				Name:    "host",
				Usage:   "The addres to run the web server on",
				Value:   "0.0.0.0",
				EnvVars: []string{"HOST"},
			},
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "Log level for the program",
				Value:   "INFO",
				EnvVars: []string{"LOG_LEVEL"},
			},
			&cli.Uint64Flag{
				Name:    "port",
				Usage:   "The port to run the web server on.",
				Value:   8080,
				EnvVars: []string{"PORT"},
				Action: func(ctx *cli.Context, v uint64) error {
					if v >= 65536 {
						return fmt.Errorf("Invalid port: %v. Must be in range[0-65535]", v)
					}
					return nil
				},
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
	var lvl = new(slog.LevelVar)
	switch cCtx.String("log-level") {
	case "DEBUG":
		lvl.Set(slog.LevelDebug)
	case "WARN":
		lvl.Set(slog.LevelWarn)
	case "ERROR":
		lvl.Set(slog.LevelError)
	default:
		lvl.Set(slog.LevelInfo)
	}

	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: lvl},
		),
	)

	database, err := db.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	logger.Debug("DB initalized")

	mux := http.NewServeMux()
	mux.Handle("/audio/", audioHandler())
	mux.Handle("/feed/", feedHandler(database))
	mux.Handle("/gen/", genFeedHandler(database, cCtx))

	address := fmt.Sprintf("%s:%d", cCtx.String("host"), cCtx.Uint64("port"))
	handler := loggingWrapper(mux, logger)
	srv := &http.Server{
		Addr:         address,
		ReadTimeout:  300 * time.Second, // for long audio returns
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler,
	}
	logger.Info("starting server", slog.String("address", address))
	return srv.ListenAndServe()
}

func feedHandler(database *sql.DB) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value("logger").(*slog.Logger)

		feedId := strings.TrimPrefix(r.URL.Path, "/feed/")
		logger = logger.With(slog.String("feed_id", feedId))

		logger.Info("Getting feed from DB")
		xml, err := db.GetFeed(ctx, database, &feedId)
		if err == sql.ErrNoRows {
			logger.Error("Feed not found in Database")
			http.Error(w, "Feed not found, please generate it.", http.StatusNotFound)
		} else if err != nil {
			logger.With(slog.String("err", fmt.Sprintf("%v", err))).Error("Something went wrong when fetching feed.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			logger.Debug("Feed found in DB")
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(*xml))
		}
	}
	return http.HandlerFunc(fn)
}

func genFeedHandler(database *sql.DB, cCtx *cli.Context) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rCtx := r.Context()
		rCtx = context.WithValue(rCtx, "url", r.URL)
		logger := rCtx.Value("logger").(*slog.Logger)

		ytPathPart := strings.TrimPrefix(r.URL.Path, "/gen/")

		logger.Info("generating feed")
		feedUrl, err := genFeed(ytPathPart, database, logger, rCtx, cCtx)
		if err != nil {
			logger.With(slog.String("err", err.Error())).Error("Something went wrong when generating feed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			logger.Debug("Feed successfully generated")
			// w.Write([]byte(*feedUrl))
			w.Write([]byte(feedUrl.String()))
		}
	}
	return http.HandlerFunc(fn)
}

func genFeed(ytPathPart string, database *sql.DB, logger *slog.Logger, rCtx context.Context, cCtx *cli.Context) (*url.URL, error) {
	base_url := cCtx.String("base-url")

	youtubeUrl := fmt.Sprintf("https://www.youtube.com/%s", ytPathPart)
	logger = logger.With(slog.String("channel_url", youtubeUrl))

	cmd := exec.Command("yt-dlp", "-J", "--playlist-items=:20", youtubeUrl)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	logger = logger.With(slog.String("yt_dlp_command", fmt.Sprintf("%v", cmd.Args)))

	err := cmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			logger = logger.With(slog.String("stderr", errb.String()))
		}
		logger.Error("failed to get channel information")
		return nil, err
	}

	var c YouTubeChannel
	err = json.Unmarshal(outb.Bytes(), &c)
	if err != nil {
		logger.Error("failed to unmarshal yt-dlp output into YouTubeChannel")
		return nil, err
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
	db.CreateFeed(rCtx, database, &c.Id, &c.Title, &c.Description, &c.Url, &feedXMLStr)

	feedURL, err := url.Parse(base_url)
	if err != nil {
		logger.Error("failed to parse feed url during construction")
		return nil, err
	}
	feedURL = feedURL.JoinPath("feed", c.Id)

	var finalURL *url.URL
	format := rCtx.Value("url").(*url.URL).Query().Get("format")
	switch format {
	case "overcast":
		queryParams := url.Values{"url": {feedURL.String()}}
		overcastURL := &url.URL{
			Scheme:   "overcast",
			Host:     "x-callback-url",
			Path:     "/add",
			RawQuery: queryParams.Encode(), // escapes "url" key automatically
		}
		finalURL = overcastURL
	default:
		finalURL = feedURL
	}

	return finalURL, nil
}

func getFeedImage(c *YouTubeChannel) string {
	for _, logo := range c.Logos {
		if logo.Preference == 1 {
			return logo.Url
		}
	}
	return "https://upload.wikimedia.org/wikipedia/commons/thumb/5/59/Minecraft_missing_texture_block.svg/1024px-Minecraft_missing_texture_block.svg.png"

}
