package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	ipfs "github.com/ipfs/kubo/core"
	"github.com/lmittmann/tint"

	iface "github.com/ipfs/kubo/core/coreapi"

	"github.com/urfave/cli/v2"
	"github.com/wetware/ww"
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
		Action:    run,
		Flags:     []cli.Flag{
			//
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

func run(c *cli.Context) error {
	// Set up IPFS
	node, err := ipfs.NewNode(c.Context, &ipfs.BuildCfg{
		Online: true,
	})
	if err != nil {
		return err
	}
	defer node.Close()

	slog.InfoContext(c.Context, "node started",
		"peer", node.PeerHost.ID())
	defer slog.InfoContext(c.Context, "node stopped",
		"peer", node.PeerHost.ID())

	api, err := iface.NewCoreAPI(node)
	if err != nil {
		return err
	}

	// Serve the default network behavior.
	return ww.Server{
		Host: node.PeerHost,
		Behavior: &ww.DefaultBehavior{
			Public: api,
		},
	}.Serve(c.Context)
}
