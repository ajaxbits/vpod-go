//go:build !integration

package podcast

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/xml"
	"net/url"
	"testing"
	"time"
	"vpod/internal/data"

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

			p, _ := New(
				tt.expected.id,
				tt.expected.Title,
				*link,
				tt.expected.Description,
			)

			gotErr := UpsertPodcast(queries, *p, tt.ctx)
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

// TestUpsertPodcast_PreservesCreatedAt verifies that upserting a feed multiple
// times does not overwrite the original created_at timestamp.
func TestUpsertPodcast_PreservesCreatedAt(t *testing.T) {
	db, queries, err := initDb()
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	feedID := "created-at-test"
	link, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	// First insert
	p1, err := New(feedID, "Original Title", *link, "Original description")
	if err != nil {
		t.Fatalf("failed to create podcast: %v", err)
	}

	if err := UpsertPodcast(queries, *p1, ctx); err != nil {
		t.Fatalf("first upsert failed: %v", err)
	}

	// Retrieve the original created_at
	var originalCreatedAt time.Time
	err = db.QueryRow("SELECT created_at FROM feeds WHERE cast(id as text) = ?", feedID).Scan(&originalCreatedAt)
	if err != nil {
		t.Fatalf("failed to query created_at: %v", err)
	}

	// Small delay to ensure time difference if created_at were overwritten
	time.Sleep(10 * time.Millisecond)

	// Second upsert with different data
	p2, err := New(feedID, "Updated Title", *link, "Updated description")
	if err != nil {
		t.Fatalf("failed to create podcast: %v", err)
	}

	if err := UpsertPodcast(queries, *p2, ctx); err != nil {
		t.Fatalf("second upsert failed: %v", err)
	}

	// Retrieve created_at after second upsert
	var afterUpsertCreatedAt time.Time
	err = db.QueryRow("SELECT created_at FROM feeds WHERE cast(id as text) = ?", feedID).Scan(&afterUpsertCreatedAt)
	if err != nil {
		t.Fatalf("failed to query created_at after upsert: %v", err)
	}

	// Verify created_at was not overwritten
	if !originalCreatedAt.Equal(afterUpsertCreatedAt) {
		t.Errorf("created_at was overwritten: original=%v, after=%v", originalCreatedAt, afterUpsertCreatedAt)
	}

	// Also verify the other fields were actually updated
	var title, description string
	err = db.QueryRow("SELECT title, description FROM feeds WHERE cast(id as text) = ?", feedID).Scan(&title, &description)
	if err != nil {
		t.Fatalf("failed to query updated fields: %v", err)
	}

	if title != "Updated Title" {
		t.Errorf("title was not updated: got %q, want %q", title, "Updated Title")
	}
	if description != "Updated description" {
		t.Errorf("description was not updated: got %q, want %q", description, "Updated description")
	}
}
