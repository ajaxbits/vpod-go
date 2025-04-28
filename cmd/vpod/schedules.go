package main

import (
	"log/slog"
	"net/url"
	"time"
	"vpod/internal/data"

	"github.com/go-co-op/gocron/v2"
)

func createUpdateJob(s gocron.Scheduler, logger *slog.Logger, baseURL *url.URL, queries *data.Queries) error {
	_, err := s.NewJob(
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
	return err
}

func createFileCullingJob(s gocron.Scheduler, logger *slog.Logger) error {
	_, err := s.NewJob(
		gocron.DurationJob(
			24*time.Hour, // TODO
		),
		gocron.NewTask(
			cullFiles,
			logger,
			"./",
		),
		gocron.WithStartAt(
			gocron.WithStartImmediately(),
		),
	)
	return err
}
