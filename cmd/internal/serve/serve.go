package serve

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/hashicorp/go-memdb"
	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/event"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/urfave/cli/v2"
	ww "github.com/wetware/go"
	"github.com/wetware/go/boot"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
)

func Command() *cli.Command {
	return &cli.Command{
		Name: "serve",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Usage:   "multi`addr` of IPFS node, or \"local\"",
				Value:   "local",
			},
			&cli.StringSliceFlag{
				Name:    "env",
				EnvVars: []string{"WW_ENV"},
			},
			&cli.BoolFlag{
				Name:    "wasm-debug",
				EnvVars: []string{"WW_WASM_DEBUG"},
				Usage:   "enable wasm debug symbols",
			},
		},
		Action: serve(),
	}
}

func serve() cli.ActionFunc {
	return func(c *cli.Context) error {
		// Set up IPFS client
		////
		ipfs, err := newIPFSClient(c)
		if err != nil {
			return err
		}

		// Set up libp2p host and DHT
		////
		h, err := libp2p.New()
		if err != nil {
			return err
		}
		defer h.Close()

		dht, err := dual.New(c.Context, h)
		if err != nil {
			return err
		}
		defer dht.Close()

		h = routedhost.Wrap(h, dht)

		// Start a multicast DNS service that searches for local
		// peers in the background.
		////
		d, err := boot.MDNS{
			Host: h,
			Handler: boot.PeerHandler{
				Peerstore:    h.Peerstore(),
				Bootstrapper: dht,
			},
		}.New()
		if err != nil {
			return err
		}
		defer d.Close()

		r := wazero.NewRuntimeWithConfig(c.Context, wazero.NewRuntimeConfig().
			WithDebugInfoEnabled(c.Bool("debug")).
			WithCloseOnContextDone(true))

		wasi, err := wasi_snapshot_preview1.Instantiate(c.Context, r)
		if err != nil {
			return err
		}
		defer wasi.Close(c.Context)

		db, err := memdb.NewMemDB(&system.Schema)
		if err != nil {
			return err
		}

		// Subscribe to events.  These will be handled
		// by env.Net.Handler.
		sub, err := h.EventBus().Subscribe([]any{
			new(event.EvtLocalAddressesUpdated),
			new(event.EvtLocalProtocolsUpdated),
			new(event.EvtLocalReachabilityChanged),
			new(event.EvtNATDeviceTypeChanged),
			new(event.EvtPeerConnectednessChanged),
			new(event.EvtPeerIdentificationCompleted),
			new(event.EvtPeerIdentificationFailed),
			new(event.EvtPeerProtocolsUpdated)})
		if err != nil {
			return err
		}
		defer sub.Close()

		return ww.Env{
			IPFS: ipfs,
			Host: h,
			Cmd: system.Cmd{
				Stdin:  stdin(c),
				Args:   c.Args().Slice(),
				Env:    c.StringSlice("env"),
				Stdout: c.App.Writer,
				Stderr: c.App.ErrWriter},
			Net: system.Net{
				Host:   h,
				Router: db,
				Handler: Handler{
					Log: slog.With(
						"peer", h.ID()),
					Sub: sub}},
			FS: system.Anchor{
				Ctx:  c.Context,
				Host: h,
				IPFS: ipfs,
			}}.Bind(c.Context, r)
	}
}

var _ system.Handler = (*Handler)(nil)

type Handler struct {
	Log *slog.Logger
	Sub event.Subscription
}

func (h Handler) ServeProc(ctx context.Context, p *proc.P) error {
	defer h.Sub.Close()

	log := h.Log.With("pid", p.String())
	log.InfoContext(ctx, "process started")

	var v any
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v = <-h.Sub.Out():
			// v was modified
		}

		// handle the event
		switch ev := v.(type) {
		case *event.EvtLocalAddressesUpdated:
			log.InfoContext(ctx, "local addresses updated")
			// ignore the event fields; they're noisy

		case *event.EvtLocalProtocolsUpdated:
			// ...

		case *event.EvtLocalReachabilityChanged:
			log.InfoContext(ctx, "local reachability changed",
				"status", ev.Reachability)

		case *event.EvtNATDeviceTypeChanged:
			log.InfoContext(ctx, "nat device type changed",
				"device", ev.NatDeviceType,
				"transport", ev.TransportProtocol)

		case *event.EvtPeerConnectednessChanged:
			log.InfoContext(ctx, "peer connection changed",
				"peer", ev.Peer,
				"status", ev.Connectedness)

		case *event.EvtPeerIdentificationCompleted:
			log.DebugContext(ctx, "peer identification completed",
				"peer", ev.Peer,
				"proto-version", ev.ProtocolVersion,
				"agent-version", ev.AgentVersion,
				"conn", ev.Conn.ID())

		case *event.EvtPeerIdentificationFailed:
			// ...

		case *event.EvtPeerProtocolsUpdated:
			// ...
		}
	}
}

func newIPFSClient(c *cli.Context) (ipfs iface.CoreAPI, err error) {
	var a ma.Multiaddr
	if s := c.String("ipfs"); s == "local" {
		ipfs, err = rpc.NewLocalApi()
	} else if a, err = ma.NewMultiaddr(s); err == nil {
		ipfs, err = rpc.NewApiWithClient(a, http.DefaultClient)
	}

	return
}

func stdin(c *cli.Context) io.Reader {
	switch r := c.App.Reader.(type) {
	case *os.File:
		info, err := r.Stat()
		if err != nil {
			panic(err)
		}

		if info.Size() <= 0 {
			break
		}

		return &io.LimitedReader{
			R: c.App.Reader,
			N: 1<<32 - 1, // max u32
		}
	}

	return &bytes.Reader{} // empty buffer
}
