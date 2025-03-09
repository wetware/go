package serve

import (
	"fmt"

	"github.com/hashicorp/go-memdb"
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
			&cli.StringSliceFlag{
				Name:    "env",
				EnvVars: []string{"WW_ENV"},
			},
			&cli.BoolFlag{
				Name:    "wasm-debug",
				EnvVars: []string{"WW_WASM_DEBUG"},
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
		Usage:  "serve a wetware process",
	}
}

func setup(*system.Env) cli.BeforeFunc {
	return func(c *cli.Context) error {
		return nil
	}
}

// serve the main event loop
func serve(env *system.Env) cli.ActionFunc {
	return func(c *cli.Context) error {
		if c.NArg() < 1 {
			return fmt.Errorf("missing required argument: path to wetware process")
		}

		// Instantiate the root process
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
