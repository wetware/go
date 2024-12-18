package client

import (
	"fmt"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/boot"
	"github.com/wetware/go/system"
)

func Command(env *system.Env) *cli.Command {
	return &cli.Command{
		Name:        "client",
		Subcommands: []*cli.Command{
			// deliver.Command(),
		},
		Action: func(c *cli.Context) error {
			// XXX:  this is a test.  remove when done.

			d, err := boot.MDNS{Env: env}.ListenAndServe()
			if err != nil {
				return err
			}
			defer d.Close()

			id, err := peer.Decode(c.Args().First())
			if err != nil {
				return fmt.Errorf("invalid peer id: %w", err)
			}

			info := peer.AddrInfo{ID: id}
			slog.InfoContext(c.Context, "connecting to peer", "peer", info.ID)

			if err := env.Host.Connect(c.Context, info); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

			return nil
		},
	}
}
