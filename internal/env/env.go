package env

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"time"
	"vpod/internal/data"
	"vpod/internal/scheduledjobs"

	"github.com/go-co-op/gocron/v2"
)

type Env struct {
	auth      AuthInfo
	BaseURL   *url.URL
	Database  *sql.DB
	Logger    *slog.Logger
	Queries   *data.Queries
	Scheduler *gocron.Scheduler
}

type AuthInfo struct {
	User string
	Pass string
}

func NewEnv(
	logLevel string,
	baseURL string,
	user string,
	pass string,
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
		auth: AuthInfo{
			User: user,
			Pass: pass,
		},
		BaseURL:   u,
		Database:  db,
		Logger:    l,
		Queries:   q,
		Scheduler: s,
	}, nil
}

func (e *Env) GetAuth() AuthInfo {
	return e.auth
}

func (e *Env) Cleanup() {
	if e.Scheduler != nil {
		s := *e.Scheduler
		s.Shutdown()
	}
	if e.Database != nil {
		e.Database.Close()
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

	if err := scheduledjobs.CreateUpdateJob(s, logger, baseURL, queries); err != nil {
		return nil, err
	}

	if err = scheduledjobs.CreateFileCullingJob(s, logger); err != nil {
		return nil, err
	}

	s.Start()
	return &s, nil
}
