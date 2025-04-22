package youtube

type Playlist struct {
	ChannelId   string `json:"channel_id"`
	ChannelName string `json:"channel"`
	ChannelUrl  string `json:"channel_url"`
	Description string
	Id          string
	Title       string
	Videos      []Video `json:"entries"`
}
