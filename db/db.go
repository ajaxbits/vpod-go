package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

func Initialize() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./podcasts.db")
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}
	return db, nil
}

func createTables(db *sql.DB) error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS Feeds (
            id BLOB PRIMARY KEY NOT NULL UNIQUE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            description TEXT,
            title TEXT NOT NULL,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            link TEXT NOT NULL,
            xml TEXT NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS Episodes (
            id BLOB PRIMARY KEY NOT NULL UNIQUE,
            audio_url TEXT NOT NULL,
            description TEXT,
            feed_id TEXT NOT NULL,
            title TEXT NOT NULL,
            FOREIGN KEY (feed_id) REFERENCES Feeds(id)
        );`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return err
		}
	}

	return nil
}
