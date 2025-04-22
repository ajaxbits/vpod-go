package podcast

import (
	"net/url"
	"testing"
	"time"
)

type want struct {
	description   string
	id            string
	lastBuildDate string
	link          string
	pubDate       string
	title         string
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		id      string
		link    url.URL
		title   string
		opts    []Option
		want    want
		wantErr bool
	}{
		{
			name:  "happy path defaults",
			id:    "my-id",
			title: "This is a podcast",
			link: url.URL{
				Scheme: "https",
				Host:   "www.google.com",
			},
			want: want{
				id:            "my-id",
				title:         "This is a podcast",
				description:   "",
				link:          "https://www.google.com",
				pubDate:       time.Now().UTC().Format(time.RFC1123Z),
				lastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
			},
			wantErr: false,
		},
		{
			name:  "with description",
			id:    "my-id",
			title: "This is a podcast",
			link: url.URL{
				Scheme: "https",
				Host:   "www.google.com",
			},
			opts: []Option{WithDescription("Some Description eh?")},
			want: want{
				id:            "my-id",
				title:         "This is a podcast",
				description:   "Some Description eh?",
				link:          "https://www.google.com",
				pubDate:       time.Now().UTC().Format(time.RFC1123Z),
				lastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
			},
			wantErr: false,
		},
		{
			name:  "with pub date",
			id:    "my-id",
			title: "This is a podcast",
			link: url.URL{
				Scheme: "https",
				Host:   "www.google.com",
			},
			opts: []Option{WithPubDate(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))},
			want: want{
				id:            "my-id",
				title:         "This is a podcast",
				description:   "",
				link:          "https://www.google.com",
				pubDate:       "Tue, 17 Nov 2009 20:34:58 +0000",
				lastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
			},
			wantErr: false,
		},
		{
			name:  "no id",
			id:    "",
			title: "This is a podcast",
			link: url.URL{
				Scheme: "https",
				Host:   "www.google.com",
			},
			wantErr: true,
		},
		{
			name:  "no title",
			id:    "my-id",
			title: "",
			link: url.URL{
				Scheme: "https",
				Host:   "www.google.com",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := New(tt.id, tt.title, tt.link, tt.opts...)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("New() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("New() succeeded unexpectedly")
			}

			if got.Id != tt.want.id {
				t.Errorf("New().Id = %v, want %v", got.Id, tt.want.id)
			}
			if got.Title != tt.want.title {
				t.Errorf("New().Title = %v, want %v", got.Title, tt.want.title)
			}
			if got.Description != tt.want.description {
				t.Errorf("New().Description = %v, want %v", got.Description, tt.want.description)
			}
			if got.PubDate != tt.want.pubDate {
				t.Errorf("New().PubDate = %v, want %v", got.PubDate, tt.want.pubDate)
			}
			if got.LastBuildDate != tt.want.lastBuildDate {
				t.Errorf("New().LastBuildDate = %v, want %v", got.LastBuildDate, tt.want.lastBuildDate)
			}

		})
	}
}
