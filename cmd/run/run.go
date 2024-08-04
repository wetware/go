package run

import (
	"net/http"

	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/thejerf/suture/v4"
	"github.com/urfave/cli/v2"
	ww "github.com/wetware/go"
	"github.com/wetware/go/util"
)

func Command() *cli.Command {
	return &cli.Command{
		Name: "run",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Usage:   "multi`addr` of IPFS node, or \"local\"",
				Value:   "local",
			},
			&cli.StringSliceFlag{
				Name:    "load",
				EnvVars: []string{"WW_LOAD"},
			},
		},
		Before: setup(),
		Action: run(),
	}
}

// Global IPFS node.  This is usually colocated on the local host,
// or available on the local network.  By default, the setup func
// will search for a local ~/.ipfs directory.
var node iface.CoreAPI

func setup() cli.BeforeFunc {
	return func(c *cli.Context) (err error) {
		var a ma.Multiaddr
		if s := c.String("ipfs"); s == "local" {
			node, err = rpc.NewLocalApi()
		} else if a, err = ma.NewMultiaddr(s); err == nil {
			node, err = rpc.NewApiWithClient(a, http.DefaultClient)
		}

		return
	}
}

func run() cli.ActionFunc {
	return func(c *cli.Context) error {
		h, err := libp2p.New()
		if err != nil {
			return err
		}
		defer h.Close()

		wetware := suture.New("ww", suture.Spec{
			EventHook: util.EventHook,
		})

		for _, ns := range c.StringSlice("load") {
			wetware.Add(ww.Config{
				NS:   ns,
				IPFS: node,
				Host: routedhost.Wrap(h, node.Routing()),
			}.Build(c.Context))
		}

		return wetware.Serve(c.Context)
	}
}
