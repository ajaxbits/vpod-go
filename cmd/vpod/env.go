package main

import (
	"context"
	"database/sql"
	"vpod/internal/data"
)

type Env struct {
	database *sql.DB
	queries  *data.Queries
}

func NewEnv() (*Env, error) {
	db, q, err := data.Initialize(context.Background())
	if err != nil {
		return nil, err
	}
	return &Env{
		database: db,
		queries:  q,
	}, nil
}
