package run

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/thejerf/suture/v4"
	"github.com/urfave/cli/v2"
	ww "github.com/wetware/go"
	"github.com/wetware/go/util"
)

// DiscoveryNotifee implements the discovery.Notifee interface
// It will be triggered when new peers are found
type DiscoveryNotifee struct {
	Ctx  context.Context
	Host host.Host
}

// HandlePeerFound adds the peer's address to the peerstore.
func (n DiscoveryNotifee) HandlePeerFound(info peer.AddrInfo) {
	n.Host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	slog.Info("discovered peer", "peer", info.ID)
}

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
			&cli.StringFlag{
				Name:    "mdns",
				EnvVars: []string{"WW_MDNS"},
				Usage:   "service tag name",
				Value:   "ww.local",
			},
		},
		Before: setup(),
		Action: run(),
	}
}

// Global IPFS ipfs.  This is usually colocated on the local host,
// or available on the local network.  By default, the setup func
// will search for a local ~/.ipfs directory.
var ipfs iface.CoreAPI

func setup() cli.BeforeFunc {
	return func(c *cli.Context) (err error) {
		var a ma.Multiaddr
		if s := c.String("ipfs"); s == "local" {
			ipfs, err = rpc.NewLocalApi()
		} else if a, err = ma.NewMultiaddr(s); err == nil {
			ipfs, err = rpc.NewApiWithClient(a, http.DefaultClient)
		}

		return
	}
}

func run() cli.ActionFunc {
	return func(c *cli.Context) error {
		h, err := libp2p.New()
		if err != nil {
			return fmt.Errorf("host: %w", err)
		}
		defer h.Close()

		// Create an mDNS service to discover peers on the local network
		service := mdns.NewMdnsService(h, c.String("mdns"), &DiscoveryNotifee{Host: h})
		if err := service.Start(); err != nil {
			return fmt.Errorf("mdns: start: %w", err)
		}
		defer service.Close()

		wetware := suture.New("ww", suture.Spec{
			EventHook: util.EventHookWithContext(c.Context),
		})

		for _, ns := range c.StringSlice("load") {
			wetware.Add(ww.Config{
				NS:   ns,
				IPFS: ipfs,
				Host: h,
			}.Build())
		}

		return wetware.Serve(c.Context)
	}
}
