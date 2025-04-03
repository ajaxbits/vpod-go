package main

import (
	"encoding/json"
	"strings"
	"time"
)

type UnixTime struct {
	time.Time
}

func (u *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	u.Time = time.Unix(timestamp, 0)
	return nil
}

type YouTubeChannel struct {
	Author      string `json:"uploader"`
	Description string
	Id          string               `json:"channel_id"`
	Logos       []YouTubeChannelLogo `json:"thumbnails"`
	Videos      []YouTubeVideo       `json:"entries"`
	Title       string
	Url         string `json:"channel_url"`
}

func (ytc *YouTubeChannel) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Author      string `json:"uploader"`
		Description string
		Entries     json.RawMessage      `json:"entries"`
		Id          string               `json:"channel_id"`
		Logos       []YouTubeChannelLogo `json:"thumbnails"`
		Title       string
		Url         string `json:"channel_url"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	var firstEntry []map[string]interface{}
	if err := json.Unmarshal(tmp.Entries, &firstEntry); err != nil {
		return err
	}
	eType, typeKeyExists := firstEntry[0]["_type"].(string)
	if typeKeyExists && eType == "playlist" {
		var playlists []YouTubePlaylist
		if err := json.Unmarshal(tmp.Entries, &playlists); err != nil {
			return err
		}

		var videosPlaylist YouTubePlaylist
		for _, playlist := range playlists {
			if strings.Contains(strings.ToLower(playlist.Title), "videos") {
				videosPlaylist = playlist
			}
		}
		ytc.Videos = videosPlaylist.Videos
	} else {
		var videos []YouTubeVideo
		if err := json.Unmarshal(tmp.Entries, &videos); err != nil {
			return err
		}
		ytc.Videos = videos
	}

	ytc.Author = tmp.Author
	ytc.Description = tmp.Description
	ytc.Id = tmp.Id
	ytc.Logos = tmp.Logos
	ytc.Title = tmp.Title
	ytc.Url = tmp.Url
	return nil
}

type YouTubeChannelLogo struct {
	Id         string
	Preference int `json:"preference,omitempty"`
	Url        string
}

type YouTubePlaylist struct {
	ChannelId   string `json:"channel_id"`
	ChannelName string `json:"channel"`
	ChannelUrl  string `json:"channel_url"`
	Description string
	Id          string
	Title       string
	Videos      []YouTubeVideo `json:"entries"`
}

type YouTubeVideo struct {
	ChannelId        string `json:"channel_id"`
	ChannelTitle     string `json:"channel"`
	ChannelUrl       string `json:"channel_url"`
	Description      string
	Duration         int64  `json:"duration"`
	DurationString   string `json:"duration_string"`
	Formats          []YouTubeVideoFormat
	Id               string
	PlaylistId       string   `json:"playlist_id"`
	ReleaseTimestamp UnixTime `json:"timestamp"`
	Thumbnail        string
	Title            string
	Url              string `json:"webpage_url"`
}

type YouTubeVideoFormat struct {
	Abr           float32
	AudioChannels int    `json:"audio_channels"`
	AudioCodec    string `json:"acodec"`
	AudioExt      string `json:"audio_ext"`
	Container     string
	Description   string `json:"format"`
	Drm           bool   `json:"has_drm"`
	Ext           string
	Language      string `json:"language"`

	// no idea what the difference between these two is
	Filesize       int64
	FilesizeApprox int64 `json:"filesize_approx"`

	Id         string `json:"format_id"`
	Protocol   string
	Resolution string
	Url        string
}
