package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v2"

	"github.com/wetware/go/cmd/ww/args"
	"github.com/wetware/go/cmd/ww/export"
	"github.com/wetware/go/cmd/ww/idgen"
	importcmd "github.com/wetware/go/cmd/ww/import"
	"github.com/wetware/go/cmd/ww/run"
	"github.com/wetware/go/cmd/ww/shell"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill)
	defer cancel()

	app := &cli.App{
		Name:      "wetware",
		Version:   "0.1.0",
		Copyright: "2020 The Wetware Project",
		Before:    setup,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:  "path",
				Value: "~/.ww",
			},
			&cli.BoolFlag{
				Name:    "json",
				EnvVars: []string{"WW_JSON"},
				Usage:   "output json logs",
			},
			&cli.StringFlag{
				Name:    "loglvl",
				EnvVars: []string{"WW_LOGLVL"},
				Value:   "info",
				Usage:   "logging level: debug, info, warn, error",
			},
		},
		Commands: []*cli.Command{
			idgen.Command(),
			run.Command(),
			shell.Command(),
			export.Command(),
			importcmd.Command(),
		},
	}

	hostArgs, guestArgs := args.SplitArgs(os.Args)
	ctx = context.WithValue(ctx, args.GuestArgs, guestArgs)

	err := app.RunContext(ctx, hostArgs)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}

func setup(c *cli.Context) (err error) {
	log := slog.New(logger(c))
	slog.SetDefault(log)
	return nil
}

func logger(c *cli.Context) slog.Handler {
	// For robots?
	if c.Bool("json") {
		return slog.NewJSONHandler(c.App.ErrWriter, &slog.HandlerOptions{
			Level: loglvl(c),
		})
	}

	// For people
	return tint.NewHandler(c.App.ErrWriter, &tint.Options{
		Level:      loglvl(c),
		TimeFormat: time.Kitchen,
	})
}

func loglvl(c *cli.Context) slog.Leveler {
	switch c.String("loglvl") {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}

	return slog.LevelInfo
}
