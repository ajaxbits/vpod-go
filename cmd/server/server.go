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

func panicHandler(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("PANIC", slog.Any("err", err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

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

	r := NewRouter()
	r.Use(middleware.NewLogging(logger))
	r.Use(panicHandler(logger))

	r.Group(func(r *Router) {
		// TODO handlerFunc
		r.Handle("GET /audio/", handlers.Audio())
		r.Handle("GET /feed/", handlers.FeedLegacy(env.queries))
		r.Handle("GET /gen/", handlers.GenFeedLegacy(cCtx, env.queries))

		r.Group(func(r *Router) {
			r.Use(middleware.NewBasicAuth(auth))

			r.HandleFunc("POST /ui/gen/", handlers.GenFeed(cCtx, env.queries))
			r.Handle("GET /ui/static/", http.StripPrefix("/ui/static/", handlers.Static()))
			r.Handle("GET /ui/", handlers.Index())
		})
	})

	address := fmt.Sprintf("%s:%d", cCtx.String("host"), cCtx.Uint64("port"))
	srv := &http.Server{
		Addr:         address,
		ReadTimeout:  300 * time.Second, // for long audio returns
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  300 * time.Second,
		Handler:      r,
	}
	logger.Info("starting server", slog.String("address", address))
	return srv.ListenAndServe()
}
