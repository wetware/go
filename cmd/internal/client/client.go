package client

import (
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/cmd/internal/client/deliver"
)

func Command() *cli.Command {
	return &cli.Command{
		Name: "client",
		Subcommands: []*cli.Command{
			deliver.Command(),
		},
	}
}
