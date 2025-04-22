package youtube

type Video struct {
	ChannelId        string `json:"channel_id"`
	ChannelTitle     string `json:"channel"`
	ChannelUrl       string `json:"channel_url"`
	Description      string
	Duration         int64  `json:"duration"`
	DurationString   string `json:"duration_string"`
	Formats          []VideoFormat
	Id               string
	PlaylistId       string   `json:"playlist_id"`
	ReleaseTimestamp UnixTime `json:"timestamp"`
	Thumbnail        string
	Title            string
	Url              string `json:"webpage_url"`
}

type VideoFormat struct {
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
