package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
	"vpod/internal/env"
	"vpod/internal/handlers"
	"vpod/internal/middleware"

	"github.com/urfave/cli/v2"
)

type CliFlags struct {
	BaseUrl string
	Port    int64
}

func main() {
	app := &cli.App{
		Name:  "vpod",
		Usage: "beware the pipeline",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "base-url",
				Usage:   "The base url for the podcast",
				EnvVars: []string{"BASE_URL"},
			},
			&cli.StringFlag{
				Name:    "host",
				Usage:   "The addres to run the web server on",
				Value:   "0.0.0.0",
				EnvVars: []string{"HOST"},
			},
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "Log level for the program",
				Value:   "INFO",
				EnvVars: []string{"LOG_LEVEL"},
			},
			&cli.StringFlag{
				Name:    "user",
				Usage:   "Username for frontend auth",
				Value:   "admin",
				EnvVars: []string{"USER"},
			},
			&cli.StringFlag{
				Name:    "password",
				Usage:   "Password for frontend auth",
				EnvVars: []string{"PASSWORD"},
			},
			&cli.Uint64Flag{
				Name:  "port",
				Usage: "The port to run the web server on.",

				Value:   8080,
				EnvVars: []string{"PORT"},
				Action: func(ctx *cli.Context, v uint64) error {
					if v >= 65536 {
						return fmt.Errorf("Invalid port: %v. Must be in range[0-65535]", v)
					}
					return nil
				},
			},
		},
		Action: func(cCtx *cli.Context) error {
			err := serve(cCtx)
			return err
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(cCtx *cli.Context) error {
	env, err := env.NewEnv(
		cCtx.String("log-level"),
		cCtx.String("base-url"),
		cCtx.String("user"),
		cCtx.String("password"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer env.Cleanup()

	logger := env.Logger
	logger.Debug("Env initalized")

	mux := http.NewServeMux()
	mux.Handle("/audio/", handlers.AudioHandler())
	mux.Handle("/feed/", handlers.FeedHandler(env.Queries))
	mux.Handle("/gen/", handlers.GenFeedHandler(cCtx, env.Queries))

	mux.Handle("/ui/gen/", middleware.BasicAuth(env, handlers.GenFeedController(cCtx, env.Queries)))
	// TODO: clean
	mux.Handle("/ui/static/", http.StripPrefix(
		"/ui/static/",
		middleware.BasicAuth(env, handlers.StaticHandler()),
	))
	mux.Handle("/ui/", middleware.BasicAuth(env, handlers.IndexHandler()))

	address := fmt.Sprintf("%s:%d", cCtx.String("host"), cCtx.Uint64("port"))
	handler := loggingWrapper(mux, logger)
	srv := &http.Server{
		Addr:         address,
		ReadTimeout:  300 * time.Second, // for long audio returns
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  300 * time.Second,
		Handler:      handler,
	}
	logger.Info("starting server", slog.String("address", address))
	return srv.ListenAndServe()
}
