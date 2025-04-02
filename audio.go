package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/eduncan911/podcast"
	"log/slog"
	"mime"
	"net/http"
	"os/exec"
	"strings"
)

type AudioMetadata struct {
	FormatId string
	VideoId  string
}

func audioHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value("logger").(*slog.Logger)
		audioPart := strings.TrimPrefix(r.URL.Path, "/audio/")
		audioParts := strings.Split(audioPart, "/") // TODO: look into SplitSeq for performance
		m := AudioMetadata{
			FormatId: audioParts[1],
			VideoId:  audioParts[0],
		}
		logger = logger.With(slog.String("audio_metadata", fmt.Sprintf("%+v", m)))
		audioFilename, err := getAudio(m, logger)
		if err != nil {
			logger.Error("Failed to get audio")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			mime.AddExtensionType(".m4a", podcast.M4A.String())
			http.ServeFile(w, r, *audioFilename)
		}
	}
	return http.HandlerFunc(fn)
}

func getAudio(m AudioMetadata, logger *slog.Logger) (*string, error) {
	youtubeUrl := fmt.Sprintf("https://www.youtube.com/watch?v=%s", m.VideoId)
	logger = logger.With(slog.String("video_url", youtubeUrl))

	cmd := exec.Command("yt-dlp", fmt.Sprintf("--format=%s", m.FormatId), "--embed-metadata", "--embed-thumbnail", "--sponsorblock-remove=sponsor", "--output=%(id)s.m4a", youtubeUrl)
	var errb bytes.Buffer
	cmd.Stderr = &errb
	logger = logger.With(slog.String("yt_dlp_command", fmt.Sprintf("%v", cmd.Args)))

	logger.Info("getting audio")
	err := cmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			logger = logger.With(slog.String("stderr", errb.String()))
		}
		logger.Error("failed to download audio from youtube",
			slog.String("err", err.Error()),
		)
		return nil, err
	}

	filename := fmt.Sprintf("%s.m4a", m.VideoId)
	return &filename, nil
}
