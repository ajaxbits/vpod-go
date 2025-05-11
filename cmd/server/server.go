package server

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"
	"vpod/internal/handlers"
	"vpod/internal/middleware"

	"github.com/urfave/cli/v2"
)

func Serve(cCtx *cli.Context) error {
	env, err := NewEnv(
		cCtx.String("log-level"),
		cCtx.String("base-url"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer env.Cleanup()

	auth := middleware.AuthInfo{
		User: cCtx.String("user"),
		Pass: cCtx.String("password"),
	}

	logger := env.logger
	logger.Debug("Env initalized")

	mainRouter := http.NewServeMux()
	mainRouter.Handle("GET /audio/", handlers.Audio())
	mainRouter.Handle("GET /feed/", handlers.FeedLegacy(env.queries))
	mainRouter.Handle("GET /gen/", handlers.GenFeedLegacy(cCtx, env.queries))

	protected := http.NewServeMux()
	protected.Handle("POST /gen/", handlers.GenFeed(cCtx, env.queries))
	protected.Handle("GET /static/", http.StripPrefix("/static/", handlers.Static()))
	protected.Handle("GET /", handlers.Index())
	mainRouter.Handle("/ui/", http.StripPrefix("/ui", middleware.BasicAuth(auth, protected)))

	address := fmt.Sprintf("%s:%d", cCtx.String("host"), cCtx.Uint64("port"))
	handler := loggingWrapper(mainRouter, logger)
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
