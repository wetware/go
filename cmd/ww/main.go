package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v2"

	"github.com/wetware/go/cmd/internal/client"
	"github.com/wetware/go/cmd/internal/export"
	"github.com/wetware/go/cmd/internal/run"
	"github.com/wetware/go/cmd/internal/serve"
	"github.com/wetware/go/system"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill)
	defer cancel()

	app := &cli.App{
		Name:      "wetware",
		Copyright: "2020 The Wetware Project",
		Before:    setup,
		// DefaultCommand: "shell",
		Flags: []cli.Flag{
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
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Usage:   "multi`addr` of IPFS node, or \"local\"",
				Value:   "local",
			},
		},
		Commands: []*cli.Command{
			run.Command(),
			serve.Command(),
			export.Command(),
			client.Command(),
		},
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}

func setup(c *cli.Context) error {
	log := slog.New(logger(c)).With(
		"version", system.Proto.Version)
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
