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
	Description string
	Id          string            `json:"channel_id"`
	Playlists   []YouTubePlaylist `json:"entries"`
	Title       string
	LastUpdated UnixTime `json:"epoch"`
	Url         string   `json:"channel_url"`
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
	Formats          []YouTubeVideoFormat
	Id               string
	PlaylistId       string   `json:"playlist_id"`
	ReleaseTimestamp UnixTime `json:"release_timestamp"` // TODO: check that this is actually a unixtime lol
	Thumbnail        string
	Title            string
	Url              string `json:"webpage_url"`
}

type YouTubeVideoFormat struct {
	AudioExt    string `json:"audio_ext"`
	Description string `json:"format"`
	Drm         bool   `json:"has_drm"`
	Ext         string
	Id          string `json:"format_id"`
	Protocol    string
	Resolution  string
	Url         string
	VideoExt    string `json:"video_ext"`
}
