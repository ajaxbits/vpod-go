package main

import (
	"encoding/json"
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
	Playlists   []YouTubePlaylist    `json:"entries"`
	Title       string
	Url         string `json:"channel_url"`
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
	Duration         int64
	DurationString   string `json:"duration_string"`
	Formats          []YouTubeVideoFormat
	Id               string
	PlaylistId       string   `json:"playlist_id"`
	ReleaseTimestamp UnixTime `json:"release_timestamp"` // TODO: check that this is actually a unixtime lol
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

	// no idea what the difference between these two is
	Filesize       int64
	FilesizeApprox int64 `json:"filesize_approx"`

	Id         string `json:"format_id"`
	Protocol   string
	Resolution string
	Url        string
}
