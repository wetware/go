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
	wetware = suture.New("ww", suture.Spec{
		EventHook:         util.EventHook,
		PassThroughPanics: true,
	})

	env system.Env
)

func Command() *cli.Command {
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
		Before: setup,
		Action: serve,
		After:  teardown,
		Usage:  "serve a wetware process",
	}
}

// serve is the main event loop for the wetware process. It:
// 1. Sets up a WebAssembly runtime environment
// 2. Loads and instantiates the process module
// 3. Initializes the routing infrastructure
// 4. Starts system services
// 5. Runs the supervisor until completion
func serve(c *cli.Context) error {
	// Validate that a process path was provided
	if c.NArg() < 1 {
		return fmt.Errorf("missing required argument: path to wetware process")
	}

	// Initialize WebAssembly runtime with debug info if enabled
	r := wazero.NewRuntimeWithConfig(c.Context, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).
		WithDebugInfoEnabled(c.Bool("wasm-debug")))
	defer r.Close(c.Context)

	// Instantiate WASI preview1 for system interface compatibility
	wasi, err := wasi_snapshot_preview1.Instantiate(c.Context, r)
	if err != nil {
		return err
	}
	defer wasi.Close(c.Context)

	// Load the WebAssembly module bytes from the specified path
	b, err := env.Load(c.Context, c.Args().First())
	if err != nil {
		return err
	}

	// Compile the WebAssembly module
	cm, err := r.CompileModule(c.Context, b)
	if err != nil {
		return err
	}
	defer cm.Close(c.Context)

	// Create a new process instance with:
	// - A unique process ID
	// - Command line arguments
	// - Standard error output
	// - A Unix-like filesystem
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

	// Initialize the software transactional memory database that will
	// store and coordinate process state
	db, err := memdb.NewMemDB(&system.Schema)
	if err != nil {
		return err
	}

	// Begin transaction to insert the root process
	init := db.Txn(true)
	if err := init.Insert("proc", p); err != nil {
		init.Abort()
		return err
	}
	init.Commit()

	// Create router for inter-process message delivery
	rt := &Router{DB: db}

	// Initialize and bind system services to the supervisor:
	// - MDNS for local network discovery
	// - P2P for distributed communication
	// - HTTP API server
	for _, s := range []suture.Service{
		&boot.MDNS{Env: &env},
		&glia.P2P{Env: &env, Router: rt},
		// &glia.Unix{Env: env, Router: rt, Path c.String("unix")},
		&glia.HTTP{Env: &env, Router: rt, ListenAddr: c.String("http")},
	} {
		wetware.Add(s)
	}

	// Log successful server startup
	env.Log().Info("server started",
		"proc", p.String())

	// Run the supervisor until context cancellation or fatal error
	return wetware.Serve(c.Context)
}
