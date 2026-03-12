package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"vpod/internal/podcast"
)

// validVideoID matches YouTube video IDs: 11 characters of [a-zA-Z0-9_-].
var validVideoID = regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)

// validFormatID matches yt-dlp format IDs: one or more digits.
var validFormatID = regexp.MustCompile(`^[0-9]+$`)

// AudioMetadata holds the parsed and validated identifiers from an audio request.
type AudioMetadata struct {
	FormatId string
	VideoId  string
}

func Audio() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value("logger").(*slog.Logger)
		audioPart := strings.TrimPrefix(r.URL.Path, "/audio/")
		audioParts := strings.Split(audioPart, "/")
		if len(audioParts) < 2 || audioParts[0] == "" || audioParts[1] == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		videoID := audioParts[0]
		formatID := audioParts[1]

		if !validVideoID.MatchString(videoID) {
			logger.Warn("rejected invalid video ID", slog.String("video_id", videoID))
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if !validFormatID.MatchString(formatID) {
			logger.Warn("rejected invalid format ID", slog.String("format_id", formatID))
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		m := AudioMetadata{
			FormatId: formatID,
			VideoId:  videoID,
		}
		logger = logger.With(slog.String("audio_metadata", fmt.Sprintf("%+v", m)))
		audioFilename, err := getAudio(m, logger)
		if err != nil {
			logger.Error("Failed to get audio")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		mime.AddExtensionType(".m4a", podcast.M4A.String())
		http.ServeFile(w, r, *audioFilename)
	}
}

func getAudio(m AudioMetadata, logger *slog.Logger) (*string, error) {
	// Serve up video quickly if it already exists.
	// filepath.Base prevents any residual path traversal.
	filename := filepath.Base(fmt.Sprintf("%s.m4a", m.VideoId))
	fileInfo, err := os.Stat(filename)
	if err == nil {
		isNonEmpty := fileInfo.Size() != 0
		isAFile := !fileInfo.IsDir()
		if isNonEmpty && isAFile {
			return &filename, nil
		}
	}

	youtubeURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", m.VideoId)
	logger = logger.With(slog.String("video_url", youtubeURL))

	cmd := exec.Command(
		"yt-dlp",
		fmt.Sprintf("--format=%s", m.FormatId),
		"--embed-metadata",
		"--embed-thumbnail",
		"--sponsorblock-remove=sponsor",
		"--output=%(id)s.m4a",
		"--", // prevent argument injection on the positional URL
		youtubeURL,
	)
	var errb bytes.Buffer
	cmd.Stderr = &errb
	logger = logger.With(slog.String("yt_dlp_command", fmt.Sprintf("%v", cmd.Args)))

	logger.Info("getting audio")
	err = cmd.Run()
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

	return &filename, nil
}
