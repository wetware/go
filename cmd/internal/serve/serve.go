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
		},
		Before: setup(env),
		Action: serve(env),
		After:  awaitSupervisorShutdown,
	}
}

func setup(env *system.Env) cli.BeforeFunc {
	return func(c *cli.Context) error {
		db, err := memdb.NewMemDB(&system.Schema)
		if err != nil {
			return err
		}
		r := &Router{DB: db}

		// Bind services to the supervisor.  These will be started
		// in the background by setup().
		////
		for _, s := range []suture.Service{
			&boot.MDNS{Env: env},
			&boot.DHT{Env: env},
			&glia.P2P{Env: env, Router: r},
			// &glia.HTTP{Env: env, Router: r},
		} {
			app.Add(s)
		}

		// Start services in the background
		errs = app.ServeBackground(c.Context)
		select {
		case err = <-errs:
		default:
		}
		return err
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
}
