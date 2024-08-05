package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v2"

	"github.com/wetware/go/cmd/internal/export"
	"github.com/wetware/go/cmd/internal/run"
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
		Commands: []*cli.Command{
			run.Command(),
			export.Command(),
		},
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}

func setup(c *cli.Context) error {
	slog.SetDefault(slog.New(tint.NewHandler(c.App.ErrWriter, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
	})))

	return nil
}
