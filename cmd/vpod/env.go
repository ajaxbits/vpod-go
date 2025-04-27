package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"time"
	"vpod/internal/data"

	"github.com/go-co-op/gocron/v2"
)

type Env struct {
	baseURL   *url.URL
	database  *sql.DB
	logger    *slog.Logger
	queries   *data.Queries
	scheduler *gocron.Scheduler
}

func NewEnv(
	logLevel string,
	baseURL string,
) (*Env, error) {
	l := newLogger(logLevel)
	if l == nil {
		return nil, errors.New("could not initalize logger")
	}

	db, q, err := data.Initialize(context.Background())
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	s, err := newScheduler(l, u, q)
	if err != nil {
		return nil, err
	}

	return &Env{
		baseURL:   u,
		database:  db,
		logger:    l,
		queries:   q,
		scheduler: s,
	}, nil
}

func (e *Env) Cleanup() {
	if e.scheduler != nil {
		s := *e.scheduler
		s.Shutdown()
	}
	if e.database != nil {
		e.database.Close()
	}
}

func newLogger(logLevel string) *slog.Logger {
	lvl := new(slog.LevelVar)
	switch logLevel {
	case "DEBUG":
		lvl.Set(slog.LevelDebug)
	case "WARN":
		lvl.Set(slog.LevelWarn)
	case "ERROR":
		lvl.Set(slog.LevelError)
	default:
		lvl.Set(slog.LevelInfo)
	}

	return slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: lvl},
		),
	)
}

func newScheduler(logger *slog.Logger, baseURL *url.URL, queries *data.Queries) (*gocron.Scheduler, error) {
	s, err := gocron.NewScheduler(
		gocron.WithLocation(time.UTC),
		gocron.WithLogger(logger),
	)
	if err != nil {
		return nil, err
	}

	_, err = s.NewJob(
		gocron.DurationJob(
			1*time.Hour, // TODO
		),
		gocron.NewTask(
			updateAll,
			logger,
			baseURL,
			queries,
		),
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // TODO: examine
	)
	if err != nil {
		return nil, err
	}

	s.Start()
	return &s, nil
}
