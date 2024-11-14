//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	ctx := context.TODO()

	app := &cli.App{
		Name:      "deliver",
		Usage:     "read message from stdn and send to `PID`",
		ArgsUsage: "<PID>",
		// Flags:  []cli.Flag{},
		Action: deliver,
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		slog.ErrorContext(ctx, "application failed",
			"reason", err)
		os.Exit(1)
	}
}

func deliver(c *cli.Context) error {
	name := c.Args().First()
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	r := io.LimitReader(c.App.Reader, 1<<32-1) // max unit32

	if n, err := io.Copy(f, r); err != nil {
		return err
	} else {
		slog.DebugContext(c.Context, "delivered message",
			"size", n,
			"dest", name)
	}

	return nil
}
