package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-merkledag"
	ipfs "github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/lmittmann/tint"
	"github.com/multiformats/go-multihash"
	"github.com/thejerf/suture/v4"

	"github.com/urfave/cli/v2"

	"github.com/wetware/ww"
	"github.com/wetware/ww/boot"
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
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "dial",
				Aliases: []string{"d"},
				EnvVars: []string{"WW_DIAL"},
			},
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
	defer slog.WarnContext(c.Context, "node stopped",
		"peer", node.PeerHost.ID())

	api, err := coreapi.NewCoreAPI(node)
	if err != nil {
		return err
	}

	// Start the event loop
	return ww.EventLoop{
		Name: node.PeerHost.ID().String(),
		Bus:  node.PeerHost.EventBus(),
		Behavior: behavior{
			DAG:       api.Dag(),
			Peerstore: node.Peerstore,
		},

		// Services are separate threads of execution that synchronize over the
		// host's event bus.  Each service implements suture.Service.
		Services: []suture.Service{
			//
			// Bootstrap services.  These will emit boot.EvtPeerFound events.
			&boot.StaticPeers{
				Bus:   node.PeerHost.EventBus(),
				Addrs: c.StringSlice("dial"),
			},
			&boot.MDNS{
				Host: node.PeerHost,
				TTL:  peerstore.AddressTTL,
			},
			&boot.ENS{
				Bus: node.PeerHost.EventBus(),
				TTL: peerstore.AddressTTL,
			},

			//
			// Another group of related services can go here...
		},
	}.Serve(c.Context)
}

type behavior struct {
	DAG       iface.APIDagService
	Peerstore peerstore.Peerstore
}

func (b behavior) OnLocalAddrsUpdated(ctx context.Context, e event.EvtLocalAddressesUpdated) {
	defer slog.DebugContext(ctx, "local addrs updated",
		"addrs", e.Current)

	data, err := e.SignedPeerRecord.Marshal()
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal peer record",
			"reason", err)
		return
	}

	n := merkledag.NodeWithData(data)
	if err := n.SetCidBuilder(cid.V1Builder{
		Codec:  cid.Raw,
		MhType: multihash.BLAKE3,
	}); err != nil {
		slog.ErrorContext(ctx, "failed to build dag node",
			"reason", err)
		return
	}

	if err := b.DAG.Add(ctx, n); err != nil {
		slog.ErrorContext(ctx, "failed to add dag node",
			"cid", n.Cid(),
			"reason", err)
		return
	}
}

func (b behavior) OnPeerFound(ctx context.Context, e boot.EvtPeerFound) {
	defer slog.DebugContext(ctx, "found peer",
		"ttl", e.TTL,
		"peer", e.Peer.ID)

	b.Peerstore.AddAddrs(e.Peer.ID, e.Peer.Addrs, e.TTL)
}
