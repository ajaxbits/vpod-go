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
				EnvVars:  []string{"BASE_URL"},
				Name:     "base-url",
				Required: true,
				Usage:    "The base url for the podcast",
			},
			&cli.StringFlag{
				EnvVars: []string{"HOST"},
				Name:    "host",
				Usage:   "The addres to run the web server on",
				Value:   "0.0.0.0",
			},
			&cli.StringFlag{
				EnvVars: []string{"LOG_LEVEL"},
				Name:    "log-level",
				Usage:   "Log level for the program",
				Value:   "INFO",
			},
			&cli.StringFlag{
				EnvVars: []string{"USER"},
				Name:    "user",
				Usage:   "Username for frontend auth",
				Value:   "admin",
			},

			&cli.StringFlag{
				Name:    "password",
				Usage:   "Password for frontend auth",
				EnvVars: []string{"PASSWORD"},
				Action: func(ctx *cli.Context, v string) error {
					if ctx.String("password-file") != "" && v != "" {
						return fmt.Errorf("Can only provide one of --password or --password-file")
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "password-file",
				Usage:   "File containing the password for frontend auth",
				EnvVars: []string{"PASSWORD_FILE"},
				Action: func(ctx *cli.Context, v string) error {
					if ctx.String("password") != "" && v != "" {
						return fmt.Errorf("Can only provide one of --password or --password-file")
					}
					return nil
				},
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
