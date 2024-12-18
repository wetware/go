package serve

import (
	"log/slog"
	"net/http"

	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/event"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/thejerf/suture/v4"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/boot"
	"github.com/wetware/go/glia"
	"github.com/wetware/go/system"
	"github.com/wetware/go/util"
)

var (
	errs <-chan error
	app  = suture.New("ww", suture.Spec{
		EventHook: util.EventHook,
	})
)

func Command(env *system.Env) *cli.Command {
	// Bind services to the supervisor.  These will be started
	// in the background by setup().
	////
	for _, s := range []suture.Service{
		&boot.MDNS{Env: env},
		&boot.DHT{Env: env},
		&glia.RPC{Env: env},
	} {
		app.Add(s)
	}

	return &cli.Command{
		Name: "serve",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Usage:   "`addr` of IPFS node (defaults: local discovery)",
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
		Before: startSupervisor,
		Action: serve(env),
		After:  awaitSupervisorShutdown,
	}
}

func startSupervisor(c *cli.Context) (err error) {
	// Start services in the background
	errs = app.ServeBackground(c.Context)
	select {
	case err = <-errs:
	default:
	}
	return
}

// awaitSupervisorShutdown waits until app closes and checks
// for an error.  If an error is found, it is returned.
func awaitSupervisorShutdown(c *cli.Context) error {
	// supervisor was started?
	if errs != nil {
		// c.Context should be closed, so supervisor should be
		// shutting down and <-errs shouldn't block.
		//
		// Expect context.Canceled on normal shutdown.
		return <-errs
	}

	return nil
}

// serve the main event loop
func serve(env *system.Env) cli.ActionFunc {
	return func(c *cli.Context) error {
		ctx := c.Context

		sub, err := env.Host.EventBus().Subscribe([]any{
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

		log := slog.With("peer", env.Host.ID())
		log.InfoContext(ctx, "wetware started")
		defer log.WarnContext(ctx, "wetware stopped")

		var v any
		for {
			select {
			case <-ctx.Done():
				return nil
			case v = <-sub.Out():
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

		// // Start a multicast DNS service that searches for local
		// // peers in the background.
		// ////
		// d, err := boot.MDNS{
		// 	NS:           "ww.local", // TODO:  make this a multiaddr, e.g. /ipfs/QmXL3jDA9XLQzZoRvwChDsFJq9QRmeS7W7vVkC9dsfWJPn
		// 	Host:         h,
		// 	Bootstrapper: dht,
		// }.New()
		// if err != nil {
		// 	return err
		// }
		// defer d.Close()

		// r := wazero.NewRuntimeWithConfig(c.Context, wazero.NewRuntimeConfig().
		// 	WithDebugInfoEnabled(c.Bool("debug")).
		// 	WithCloseOnContextDone(true))

		// wasi, err := wasi_snapshot_preview1.Instantiate(c.Context, r)
		// if err != nil {
		// 	return err
		// }
		// defer wasi.Close(c.Context)

		// db, err := memdb.NewMemDB(&system.Schema)
		// if err != nil {
		// 	return err
		// }

		// // Subscribe to events.  These will be handled
		// // by env.Net.Handler.
		// sub, err := h.EventBus().Subscribe([]any{
		// 	new(event.EvtLocalAddressesUpdated),
		// 	new(event.EvtLocalProtocolsUpdated),
		// 	new(event.EvtLocalReachabilityChanged),
		// 	new(event.EvtNATDeviceTypeChanged),
		// 	new(event.EvtPeerConnectednessChanged),
		// 	new(event.EvtPeerIdentificationCompleted),
		// 	new(event.EvtPeerIdentificationFailed),
		// 	new(event.EvtPeerProtocolsUpdated)})
		// if err != nil {
		// 	return err
		// }
		// defer sub.Close()

		// return ww.Env{
		// 	IPFS: ipfs,
		// 	Host: h,
		// 	Cmd: system.Cmd{
		// 		Stdin:  stdin(c),
		// 		Args:   c.Args().Slice(),
		// 		Env:    c.StringSlice("env"),
		// 		Stdout: c.App.Writer,
		// 		Stderr: c.App.ErrWriter},
		// 	Net: system.Net{
		// 		Host:   h,
		// 		Router: db,
		// 		Handler: Handler{
		// 			Log: slog.With(
		// 				"peer", h.ID()),
		// 			Sub: sub}},
		// 	FS: system.Anchor{
		// 		Ctx:  c.Context,
		// 		Host: h,
		// 		IPFS: ipfs,
		// 	}}.Bind(c.Context, r)
	}
}

func newIPFSClient(c *cli.Context) (iface.CoreAPI, error) {
	if !c.IsSet("ipfs") {
		return rpc.NewLocalApi()
	}

	a, err := ma.NewMultiaddr(c.String("ipfs"))
	if err != nil {
		return nil, err
	}

	return rpc.NewApiWithClient(a, http.DefaultClient)
}

// var _ system.Handler = (*Handler)(nil)

// type Handler struct {
// 	Log *slog.Logger
// 	Sub event.Subscription
// }

// func (h Handler) ServeProc(ctx context.Context, p *proc.P) error {
// 	defer h.Sub.Close()

// 	log := h.Log.With("pid", p.String())
// 	log.InfoContext(ctx, "process started")

// 	var v any
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		case v = <-h.Sub.Out():
// 			// v was modified
// 		}

// 		// handle the event
// 		switch ev := v.(type) {
// 		case *event.EvtLocalAddressesUpdated:
// 			log.InfoContext(ctx, "local addresses updated")
// 			// ignore the event fields; they're noisy

// 		case *event.EvtLocalProtocolsUpdated:
// 			// ...

// 		case *event.EvtLocalReachabilityChanged:
// 			log.InfoContext(ctx, "local reachability changed",
// 				"status", ev.Reachability)

// 		case *event.EvtNATDeviceTypeChanged:
// 			log.InfoContext(ctx, "nat device type changed",
// 				"device", ev.NatDeviceType,
// 				"transport", ev.TransportProtocol)

// 		case *event.EvtPeerConnectednessChanged:
// 			log.InfoContext(ctx, "peer connection changed",
// 				"peer", ev.Peer,
// 				"status", ev.Connectedness)

// 		case *event.EvtPeerIdentificationCompleted:
// 			log.DebugContext(ctx, "peer identification completed",
// 				"peer", ev.Peer,
// 				"proto-version", ev.ProtocolVersion,
// 				"agent-version", ev.AgentVersion,
// 				"conn", ev.Conn.ID())

// 		case *event.EvtPeerIdentificationFailed:
// 			// ...

// 		case *event.EvtPeerProtocolsUpdated:
// 			// ...
// 		}
// 	}
// }

// func stdin(c *cli.Context) io.Reader {
// 	switch r := c.App.Reader.(type) {
// 	case *os.File:
// 		info, err := r.Stat()
// 		if err != nil {
// 			panic(err)
// 		}

// 		if info.Size() <= 0 {
// 			break
// 		}

// 		return &io.LimitedReader{
// 			R: c.App.Reader,
// 			N: 1<<32 - 1, // max u32
// 		}
// 	}

// 	return &bytes.Reader{} // empty buffer
// }
