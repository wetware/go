package serve

import (
	"log/slog"

	"github.com/hashicorp/go-memdb"
	"github.com/libp2p/go-libp2p/core/event"
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
			&cli.StringFlag{
				Name:    "http",
				EnvVars: []string{"WW_HTTP"},
				Usage:   "bind API server to `HOST`:`PORT`",
				Value:   "localhost:2080",
			},
		},
		Before: setup(env),
		Action: serve(env),
		After:  awaitSupervisorShutdown,
	}
}

func setup(env *system.Env) cli.BeforeFunc {
	return func(c *cli.Context) error {
		// Initialize STM and construct the router.
		////
		db, err := memdb.NewMemDB(&system.Schema)
		if err != nil {
			return err
		}
		r := &Router{DB: db}

		// Bind services to the supervisor.
		////
		p2p := glia.P2P{Env: env, Router: r} // core p2p service
		for _, s := range []suture.Service{
			p2p,
			// &glia.Unix{P2P: p2p, Path c.String("unix")},
			&glia.HTTP{P2P: p2p, ListenAddr: c.String("http")},
			&boot.MDNS{Env: env /*NS: c.String("ns")*/},
		} {
			app.Add(s)
		}

		// Run the supervisor
		////
		return app.Serve(c.Context)
	}
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
				log.InfoContext(ctx, "local protocols updated")

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
				log.WarnContext(ctx, "peer identification failed",
					"peer", ev.Peer,
					"reason", ev.Reason)

			case *event.EvtPeerProtocolsUpdated:
				log.InfoContext(ctx, "peer protocols updated",
					"peer", ev.Peer)

			}
		}
	}
}
