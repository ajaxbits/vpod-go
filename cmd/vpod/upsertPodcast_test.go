//go:build !integration

package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/xml"
	"net/url"
	"testing"
	"time"
	"vpod/internal/data"
	"vpod/internal/podcast"

	_ "github.com/mattn/go-sqlite3"
)

func initDb() (*sql.DB, *data.Queries, error) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		return nil, nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// create tables
	if _, err := db.ExecContext(context.Background(), data.DDL); err != nil {
		return nil, nil, err
	}

	queries := data.New(db)
	return db, queries, nil
}

type TestData struct {
	Description string `xml:"description"`
	id          string
	Link        string `xml:"link"`
	Title       string `xml:"title"`
}

func Test_upsertPodcast(t *testing.T) {
	db, queries, _ := initDb()
	ctx := context.Background()
	env := Env{
		database: db,
		queries:  queries,
	}

	tests := []struct {
		name     string
		expected TestData
		ctx      context.Context
		wantErr  bool
	}{
		{
			name: "happy path",
			expected: TestData{
				Description: "A test podcast",
				id:          "todo-test",
				Link:        "https://www.google.com",
				Title:       "A test podcast",
			},
			ctx:     ctx,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			link, err := url.Parse(tt.expected.Link)
			if err != nil {
				t.Errorf("failed: %v", err)
			}

			p, _ := podcast.New(
				tt.expected.id,
				tt.expected.Title,
				*link,
				tt.expected.Description,
			)

			gotErr := upsertPodcast(env.queries, *p, tt.ctx)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("failed: %v", gotErr)
				}
			}
			if tt.wantErr {
				t.Fatal("succeeded unexpectedly")
			}

			var (
				got    TestData
				gotXML string
			)
			err = db.QueryRow("select description, id, link, title, xml from feeds where cast(id as text) = ?", "todo-test").Scan(
				&got.Description,
				&got.id,
				&got.Link,
				&got.Title,
				&gotXML,
			)
			if err != nil {
				t.Errorf("failed: %v", err)
			}

			if got != tt.expected {
				t.Fatal("upsertPodcast() did not insert the right data")
			}

			var gotXMLData struct {
				Channel TestData `xml:"channel"`
			}
			gotErr = xml.Unmarshal([]byte(gotXML), &gotXMLData)
			if err != nil {
				t.Errorf("failed: %v", err)
			}

			gotXMLData.Channel.id = tt.expected.id
			if gotXMLData.Channel != tt.expected {
				t.Fatal("upsertPodcast() did not insert the right xml data into the DB, but got everything else right")
			}
			return
		})
	}
}
