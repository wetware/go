package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/ipfs/go-cid"
	ipfs "github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/event"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/lmittmann/tint"
	"github.com/thejerf/suture/v4"
	"golang.org/x/sync/semaphore"

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
			&cli.StringSliceFlag{
				Name:    "mdns",
				EnvVars: []string{"WW_MDNS"},
			},
			&cli.StringFlag{
				Name:    "root",
				EnvVars: []string{"WW_ROOT"},
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
	root, err := cid.Decode(c.String("root"))
	if err != nil {
		return err
	}

	rh := routedhost.Wrap(node.PeerHost, node.DHT)
	return ww.EventLoop{
		Name: node.PeerHost.ID().String(),
		Bus:  node.PeerHost.EventBus(),
		Behavior: &behavior{
			API: api,
		},

		// Services are separate threads of execution that synchronize over the
		// host's event bus.  Each service implements suture.Service.
		Services: []suture.Service{
			//
			// Bootstrap services.  These will emit boot.EvtPeerFound events.
			&boot.StaticPeers{
				Host:  rh,
				Addrs: c.StringSlice("dial"),
			},
			&boot.MDNS{
				Host:        rh,
				ServiceName: c.String("mdns"),
			},
			&boot.IPFS{
				Host: rh,
				API:  api,
				CID:  root,
			},

			//
			// Another group of related services can go here...
		},
	}.Serve(c.Context)
}

type behavior struct {
	API iface.CoreAPI

	sem *semaphore.Weighted
}

func (b *behavior) OnLocalAddrsUpdated(ctx context.Context, e event.EvtLocalAddressesUpdated) {
	defer slog.DebugContext(ctx, "local addrs updated",
		"addrs", e.Current)
}

func (b *behavior) OnPeerFound(ctx context.Context, e boot.EvtPeerFound) {
	defer slog.DebugContext(ctx, "found peer",
		"peer", e.Peer.ID)

	// We're using a semaphore as a crude rate-limiting mechanism.
	// All event handling is strictly single-threaded, so this will
	// eventually block.  At scale, we'll need
	if b.sem == nil {
		b.sem = semaphore.NewWeighted(16)
	}

	sem := b.sem
	go func() {
		// Check with the semaphore to see if a slot is available.
		// If not, just drop the event.
		if !sem.TryAcquire(1) {
			return
		}
		defer sem.Release(1)

		// To keep things moving along, we'll set a timeout on
		// the connection attempt.
		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		// Connect to the peer.  If the connection attempt is
		// successful, the peer will be added to the local store.
		if err := b.API.Swarm().Connect(ctx, e.Peer); err != nil {
			slog.DebugContext(ctx, "peer connection failed",
				"peer", e.Peer.ID,
				"addrs", e.Peer.Addrs)
		}

		slog.InfoContext(ctx, "connected to peer",
			"peer", e.Peer.ID)
	}()

	slog.DebugContext(ctx, "found peer",
		"peer", e.Peer.ID)
}
