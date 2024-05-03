package run

import (
	"github.com/ipfs/kubo/client/rpc"
	"github.com/libp2p/go-libp2p"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/thejerf/suture/v4"
	"github.com/urfave/cli/v2"
	ww "github.com/wetware/go"
	"github.com/wetware/go/util"
)

func Command() *cli.Command {
	return &cli.Command{
		Name: "run",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "load",
				EnvVars: []string{"WW_LOAD"},
			},
		},
		Action: run,
	}
}

func run(c *cli.Context) error {
	node, err := rpc.NewLocalApi()
	if err != nil {
		return err
	}

	h, err := libp2p.New()
	if err != nil {
		return err
	}

	wetware := suture.New("ww", suture.Spec{
		EventHook: util.EventHook,
	})

	for _, s := range c.StringSlice("load") {
		wetware.Add(&ww.Cluster{
			Name: s,
			IPFS: node,
			Host: routedhost.Wrap(h, node.Routing()),
		})
	}

	return wetware.Serve(c.Context)
}
