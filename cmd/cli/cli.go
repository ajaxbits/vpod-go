package cli

import (
	"fmt"
	"vpod/cmd/server"

	"github.com/urfave/cli/v2"
)

func NewApp() *cli.App {
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
			err := server.Serve(cCtx)
			return err
		},
	}
	return app
}
