package main

import (
	"fmt"
	"github.com/eduncan911/podcast"
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
		audioPart := strings.TrimPrefix(r.URL.Path, "/audio/")
		audioParts := strings.Split(audioPart, "/") // TODO: look into SplitSeq for performance
		m := AudioMetadata{
			FormatId: audioParts[1],
			VideoId:  audioParts[0],
		}
		audioFilename, err := getAudio(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		mime.AddExtensionType(".m4a", podcast.M4A.String())
		http.ServeFile(w, r, *audioFilename)
	}
	return http.HandlerFunc(fn)
}

func getAudio(m AudioMetadata) (*string, error) {
	youtubeUrl := fmt.Sprintf("https://www.youtube.com/watch?v=%s", m.VideoId)
	cmd := exec.Command("yt-dlp", fmt.Sprintf("--format=%s", m.FormatId), "--embed-metadata", "--embed-thumbnail", "--sponsorblock-remove=sponsor", "--output=%(id)s.m4a", youtubeUrl)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s.m4a", m.VideoId)
	return &filename, nil
}
