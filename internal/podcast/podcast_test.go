package podcast

import (
	"fmt"
	"net/url"
	"testing"
	"time"
	"vpod/internal/youtube"

	"github.com/eduncan911/podcast"
)

type want struct {
	description   string
	id            string
	lastBuildDate string
	link          string
	pubDate       string
	title         string
}

func formatIAuthor(author string) string {
	return fmt.Sprintf("no_email_provided (%s)", author)
}

func TestFromChannel(t *testing.T) {
	// Set up a fixed time to use for test comparisons
	fixedTime := time.Date(2023, 5, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		c       youtube.Channel
		baseURL url.URL
		opts    []Option
		want    *Podcast
		wantErr bool
	}{
		{
			name: "happy path - basic channel conversion",
			c: youtube.Channel{
				Id:          "test-channel-id",
				Title:       "Test Channel - Videos",
				Description: "This is a test channel description",
				Author:      "Test Author",
				URL: url.URL{
					Scheme: "https",
					Host:   "youtube.com",
					Path:   "/channel/test-channel-id",
				},
				Videos: []youtube.Video{
					{
						Id:               "video1",
						Title:            "Test Video 1",
						Description:      "This is test video 1",
						Url:              "https://youtube.com/watch?v=video1",
						Thumbnail:        "https://img.youtube.com/vi/video1/maxresdefault.jpg",
						Duration:         300, // 5 minutes
						ReleaseTimestamp: youtube.UnixTime{Time: fixedTime},
						Formats: []youtube.VideoFormat{
							{
								Id:         "140",
								Resolution: "audio only",
								AudioExt:   "m4a",
								Language:   "en-US",
								Filesize:   5000000,
								Drm:        false,
							},
						},
					},
				},
			},
			baseURL: url.URL{
				Scheme: "https",
				Host:   "example.com",
			},
			opts: []Option{WithPubDate(fixedTime)},
			want: &Podcast{
				Id: "test-channel-id",
				Podcast: &podcast.Podcast{
					Title:         "Test Channel",
					Description:   "This is a test channel description",
					Link:          "https://youtube.com/channel/test-channel-id",
					PubDate:       fixedTime.Format(time.RFC1123Z),
					LastBuildDate: fixedTime.Format(time.RFC1123Z),
					IAuthor:       formatIAuthor("Test Author"),
					Image: &podcast.Image{
						URL: "", // Will be set by the function
					},
					IExplicit: "no",
					IBlock:    "Yes",
					Generator: "vpod",
					ISubtitle: "This is a test channel description",
					ISummary:  &podcast.ISummary{Text: "This is a test channel description"},
					Items: []*Item{
						{
							Title:       "Test Video 1",
							Description: "This is test video 1",
							Link:        "https://youtube.com/watch?v=video1",
							PubDate:     &fixedTime,
							Enclosure: &podcast.Enclosure{
								URL:    "https://example.com/audio/video1/140",
								Type:   podcast.M4A,
								Length: 5000000,
							},
							IDuration: "5:00",
							IImage: &podcast.IImage{
								HREF: "https://img.youtube.com/vi/video1/maxresdefault.jpg",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "channel with no videos",
			c: youtube.Channel{
				Id:          "empty-channel",
				Title:       "Empty Channel - Videos",
				Description: "This channel has no videos",
				Author:      "Empty Author",
				URL: url.URL{
					Scheme: "https",
					Host:   "youtube.com",
					Path:   "/channel/empty-channel",
				},
				Videos: []youtube.Video{},
			},
			baseURL: url.URL{
				Scheme: "https",
				Host:   "example.com",
			},
			opts: []Option{WithPubDate(fixedTime)},
			want: &Podcast{
				Id: "empty-channel",
				Podcast: &podcast.Podcast{
					Title:         "Empty Channel",
					Description:   "This channel has no videos",
					Language:      "en",
					Link:          "https://youtube.com/channel/empty-channel",
					PubDate:       fixedTime.Format(time.RFC1123Z),
					LastBuildDate: fixedTime.Format(time.RFC1123Z),
					IAuthor:       formatIAuthor("Empty Author"),
					IExplicit:     "no",
					IBlock:        "Yes",
					Generator:     "vpod",
					ISubtitle:     "This channel has no videos",
					ISummary: &podcast.ISummary{
						Text: "This channel has no videos",
					},
					Items: []*Item{},
				},
			},
			wantErr: false,
		},
		{
			name: "channel with no description",
			c: youtube.Channel{
				Id:          "test-channel-id",
				Title:       "Test Channel - Videos",
				Description: "",
				Author:      "Test Author",
				URL: url.URL{
					Scheme: "https",
					Host:   "youtube.com",
					Path:   "/channel/test-channel-id",
				},
				Videos: []youtube.Video{
					{
						Id:               "video1",
						Title:            "Test Video 1",
						Description:      "This is test video 1",
						Url:              "https://youtube.com/watch?v=video1",
						Thumbnail:        "https://img.youtube.com/vi/video1/maxresdefault.jpg",
						Duration:         300, // 5 minutes
						ReleaseTimestamp: youtube.UnixTime{Time: fixedTime},
						Formats: []youtube.VideoFormat{
							{
								Id:         "140",
								Resolution: "audio only",
								AudioExt:   "m4a",
								Language:   "en-US",
								Filesize:   5000000,
								Drm:        false,
							},
						},
					},
				},
			},
			baseURL: url.URL{
				Scheme: "https",
				Host:   "example.com",
			},
			opts: []Option{WithPubDate(fixedTime)},
			want: &Podcast{
				Id: "test-channel-id",
				Podcast: &podcast.Podcast{
					Title:         "Test Channel",
					Description:   "no description provided",
					Link:          "https://youtube.com/channel/test-channel-id",
					PubDate:       fixedTime.Format(time.RFC1123Z),
					LastBuildDate: fixedTime.Format(time.RFC1123Z),
					IAuthor:       formatIAuthor("Test Author"),
					Image: &podcast.Image{
						URL: "", // Will be set by the function
					},
					IExplicit: "no",
					IBlock:    "Yes",
					Generator: "vpod",
					ISubtitle: "This is a test channel description",
					ISummary:  &podcast.ISummary{Text: "This is a test channel description"},
					Items: []*Item{
						{
							Title:       "Test Video 1",
							Description: "This is test video 1",
							Link:        "https://youtube.com/watch?v=video1",
							PubDate:     &fixedTime,
							Enclosure: &podcast.Enclosure{
								URL:    "https://example.com/audio/video1/140",
								Type:   podcast.M4A,
								Length: 5000000,
							},
							IDuration: "5:00",
							IImage: &podcast.IImage{
								HREF: "https://img.youtube.com/vi/video1/maxresdefault.jpg",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "channel with non-conforming videos",
			c: youtube.Channel{
				Id:          "non-conforming-channel",
				Title:       "Test Channel - Videos",
				Description: "This channel has non-conforming videos",
				Author:      "Test Author",
				URL: url.URL{
					Scheme: "https",
					Host:   "youtube.com",
					Path:   "/channel/non-conforming-channel",
				},
				Videos: []youtube.Video{
					{
						Id:               "video2",
						Title:            "Test Video 2",
						Description:      "This is test video 2",
						Url:              "https://youtube.com/watch?v=video2",
						Thumbnail:        "https://img.youtube.com/vi/video2/maxresdefault.jpg",
						Duration:         600, // 10 minutes
						ReleaseTimestamp: youtube.UnixTime{Time: fixedTime},
						Formats: []youtube.VideoFormat{
							{
								Id:         "141",
								Resolution: "audio only",
								AudioExt:   "mp3", // Not m4a
								Language:   "en-US",
								Filesize:   10000000,
								Drm:        false,
							},
							{
								Id:         "142",
								Resolution: "audio only",
								AudioExt:   "m4a",
								Language:   "fr-FR", // Not English
								Filesize:   9000000,
								Drm:        false,
							},
							{
								Id:         "143",
								Resolution: "audio only",
								AudioExt:   "m4a",
								Language:   "en-US",
								Filesize:   8000000,
								Drm:        true, // Has DRM
							},
							{
								Id:         "144drc",
								Resolution: "audio only",
								AudioExt:   "m4a",
								Language:   "en-US",
								Filesize:   7000000,
								Drm:        false,
							},
							{
								Id:         "145",
								Resolution: "720p", // Not audio only
								AudioExt:   "m4a",
								Language:   "en-US",
								Filesize:   15000000,
								Drm:        false,
							},
						},
					},
				},
			},
			baseURL: url.URL{
				Scheme: "https",
				Host:   "example.com",
			},
			opts: []Option{WithPubDate(fixedTime)},
			want: &Podcast{
				Id: "non-conforming-channel",
				Podcast: &podcast.Podcast{
					Title:         "Test Channel",
					Description:   "This channel has non-conforming videos",
					Language:      "en",
					Link:          "https://youtube.com/channel/non-conforming-channel",
					PubDate:       fixedTime.Format(time.RFC1123Z),
					LastBuildDate: fixedTime.Format(time.RFC1123Z),
					IAuthor:       formatIAuthor("Test Author"),
					IExplicit:     "no",
					IBlock:        "Yes",
					Generator:     "vpod",
					ISubtitle:     "This channel has non-conforming videos",
					ISummary:      &podcast.ISummary{Text: "This channel has non-conforming videos"},
					Items:         []*Item{}, // No items because all formats don't meet criteria
				},
			},
			wantErr: false,
		},
		{
			name: "error in New function",
			c: youtube.Channel{
				Id:          "", // Empty ID will cause New() to return an error
				Title:       "Error Channel - Videos",
				Description: "This channel will cause an error",
				URL: url.URL{
					Scheme: "https",
					Host:   "youtube.com",
					Path:   "/channel/error-channel",
				},
			},
			baseURL: url.URL{
				Scheme: "https",
				Host:   "example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		// t.Parallel()
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, gotErr := FromChannel(tt.c, tt.baseURL, tt.opts...)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("FromChannel() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("FromChannel() succeeded unexpectedly")
			}

			// Basic fields
			if got.Id != tt.want.Id {
				t.Errorf("FromChannel() Id = %v, want %v", got.Id, tt.want.Id)
			}

			if got.Title != tt.want.Title {
				t.Errorf("FromChannel() Title = %v, want %v", got.Title, tt.want.Title)
			}

			if got.Description != tt.want.Description {
				t.Errorf("FromChannel() Description = %v, want %v", got.Description, tt.want.Description)
			}

			if got.IAuthor != tt.want.IAuthor {
				t.Errorf("FromChannel() IAuthor = %v, want %v", got.IAuthor, tt.want.IAuthor)
			}

			if got.IExplicit != tt.want.IExplicit {
				t.Errorf("FromChannel() IExplicit = %v, want %v", got.IExplicit, tt.want.IExplicit)
			}

			if got.IBlock != tt.want.IBlock {
				t.Errorf("FromChannel() IBlock = %v, want %v", got.IBlock, tt.want.IBlock)
			}

			if got.Generator != tt.want.Generator {
				t.Errorf("FromChannel() Generator = %v, want %v", got.Generator, tt.want.Generator)
			}

			// Check items length
			if len(got.Items) != len(tt.want.Items) {
				t.Errorf("FromChannel() Items length = %v, want %v", len(got.Items), len(tt.want.Items))
			}

			// Check specific items if present
			if len(got.Items) > 0 && len(tt.want.Items) > 0 {
				for i, item := range got.Items {
					wantItem := tt.want.Items[i]

					if item.Title != wantItem.Title {
						t.Errorf("Item[%d].Title = %v, want %v", i, item.Title, wantItem.Title)
					}

					if item.Description != wantItem.Description {
						t.Errorf("Item[%d].Description = %v, want %v", i, item.Description, wantItem.Description)
					}

					if item.Link != wantItem.Link {
						t.Errorf("Item[%d].Link = %v, want %v", i, item.Link, wantItem.Link)
					}

					if *item.PubDate != *wantItem.PubDate {
						t.Errorf("Item[%d].PubDate = %v, want %v", i, item.PubDate, wantItem.PubDate)
					}

					if item.IDuration != wantItem.IDuration {
						t.Errorf("Item[%d].IDuration = %v, want %v", i, item.IDuration, wantItem.IDuration)
					}

					if *item.IImage != *wantItem.IImage {
						t.Errorf("Item[%d].IImage = %v, want %v", i, item.IImage, wantItem.IImage)
					}

					if item.Enclosure.URL != wantItem.Enclosure.URL {
						t.Errorf("Item[%d].Enclosure.URL = %v, want %v", i, item.Enclosure.URL, wantItem.Enclosure.URL)
					}

					if item.Enclosure.Type != wantItem.Enclosure.Type {
						t.Errorf("Item[%d].Enclosure.Type = %v, want %v", i, item.Enclosure.Type, wantItem.Enclosure.Type)
					}

					if item.Enclosure.Length != wantItem.Enclosure.Length {
						t.Errorf("Item[%d].Enclosure.Length = %v, want %v", i, item.Enclosure.Length, wantItem.Enclosure.Length)
					}
				}
			}
		})
	}
}
