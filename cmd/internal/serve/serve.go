package serve

import (
	"log/slog"

	"github.com/hashicorp/go-memdb"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/thejerf/suture/v4"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/boot"
	"github.com/wetware/go/glia"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
	"github.com/wetware/go/util"
)

var (
	app = suture.New("ww", suture.Spec{
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
	}
}

func setup(env *system.Env) cli.BeforeFunc {
	return func(c *cli.Context) error {
		// Intantiate the root process
		////
		r := wazero.NewRuntimeWithConfig(c.Context, wazero.NewRuntimeConfig().
			WithCloseOnContextDone(true).
			WithDebugInfoEnabled(c.Bool("wasm-debug")))
		defer r.Close(c.Context)

		wasi, err := wasi_snapshot_preview1.Instantiate(c.Context, r)
		if err != nil {
			return err
		}
		defer wasi.Close(c.Context)

		b, err := env.Load(c.Context, c.Args().First())
		if err != nil {
			return err
		}
		cm, err := r.CompileModule(c.Context, b)
		if err != nil {
			return err
		}
		defer cm.Close(c.Context)

		p, err := proc.Command{
			PID:    proc.NewPID(),
			Args:   c.Args().Slice(),
			Stderr: c.App.ErrWriter,
			FS:     env.NewUnixFS(c.Context),
		}.Instantiate(c.Context, r, cm)
		if err != nil {
			return err
		}
		defer p.Close(c.Context)

		// Initialize STM router and insert the root process
		////
		db, err := memdb.NewMemDB(&system.Schema) // provides STM
		if err != nil {
			return err
		}
		// Initialized an in-memory database that provides software-
		// transactional-memory (STM) semantics for us.  This gives
		// us flexibility to read/write multiple processes atomically.
		//
		// We add p to the "proc" table.
		init := db.Txn(true)
		if err := init.Insert("proc", p); err != nil {
			init.Abort()
			return err
		}
		init.Commit()
		rt := &Router{DB: db} // message-routing interface; can route messages locally

		// Bind services to the supervisor.
		////
		for _, s := range []suture.Service{
			&boot.MDNS{Env: env},
			&glia.P2P{Env: env, Router: rt},
			// &glia.Unix{Env: env, Router: rt, Path c.String("unix")},
			&glia.HTTP{Env: env, Router: rt, ListenAddr: c.String("http")},
		} {
			app.Add(s)
		}

		env.Log().Info("server started",
			"proc", p.String())

		// Run the supervisor
		////
		return app.Serve(c.Context)
	}
}

// func bind(db *memdb.MemDB, p *proc.P) (err error) {
// 	tx := db.Txn(true)
// 	defer tx.Commit()

// 	if err = tx.Insert("proc", p); err != nil {
// 		tx.Abort()
// 	}
// 	return
// }

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
