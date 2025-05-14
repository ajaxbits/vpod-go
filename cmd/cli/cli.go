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
			&cli.BoolFlag{
				EnvVars: []string{"NO_AUTH"},
				Name:    "no-auth",
				Usage:   "Deactivate authentication for the frontend",
				Value:   false,
			},
			&cli.StringFlag{
				EnvVars: []string{"USER"},
				Name:    "user",
				Usage:   "Username for frontend auth",
				Value:   "admin",
				Action: func(ctx *cli.Context, val string) error {
					if val == "" {
						return fmt.Errorf("User cannot be empty")
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "password",
				Usage:   "Password for frontend auth",
				EnvVars: []string{"PASSWORD"},
			},
			&cli.StringFlag{
				Name:    "password-file",
				Usage:   "File containing the password for frontend auth",
				EnvVars: []string{"PASSWORD_FILE"},
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
		Before: func(ctx *cli.Context) error {
			authEnabled := !ctx.Bool("no-auth")
			passwordVal := ctx.String("password")
			passwordFileVal := ctx.String("password-file")
			userVal := ctx.String("user")

			if authEnabled {
				userEmpty := userVal == ""
				userProvidedButNoPass := userVal != "" && passwordFileVal == "" && passwordVal == ""
				bothPassFlagsSet := passwordVal != "" && passwordFileVal != ""

				if userEmpty {
					return fmt.Errorf("When auth is enabled, user cannot be empty.")
				}
				if userProvidedButNoPass {
					return fmt.Errorf("Password is required when auth enabled and user specified.")
				}
				if bothPassFlagsSet {
					return fmt.Errorf("Cannot set both a password and a password-file.")
				}
			}

			return nil
		},
		Action: func(cCtx *cli.Context) error {
			err := server.Serve(cCtx)
			return err
		},
	}
	return app
}
