package data

import (
	"context"
	"database/sql"
	_ "embed"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var DDL string

func Initialize(ctx context.Context) (*sql.DB, *Queries, error) {
	db, err := sql.Open("sqlite3", "./podcasts.db")
	if err != nil {
		return nil, nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if _, err := db.ExecContext(ctx, DDL); err != nil {
		return nil, nil, err
	}

	queries := New(db)
	return db, queries, nil
}
