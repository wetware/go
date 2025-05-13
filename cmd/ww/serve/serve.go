package serve

import (
	"context"
	"io"
	"os"

	"github.com/hashicorp/go-memdb"
	"github.com/ipfs/boxo/path"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
	"github.com/thejerf/suture/v4"
	"github.com/urfave/cli/v2"
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
			&cli.StringFlag{
				Name:    "ns",
				EnvVars: []string{"WW_NS"},
				Value:   "ww",
				Usage:   "cluster namespace",
			},
			&cli.StringSliceFlag{
				Name:    "dial",
				EnvVars: []string{"WW_DIAL"},
				Aliases: []string{"d"},
				Usage:   "peer addr to dial",
			},
			&cli.StringSliceFlag{
				Name:    "listen",
				EnvVars: []string{"WW_LISTEN"},
				Aliases: []string{"l"},
				Usage:   "multiaddr to listen on",
			},
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Usage:   "multi`addr` of IPFS node, or \"local\"",
				Value:   "local",
			},
			&cli.StringFlag{
				Name:    "privkey",
				Aliases: []string{"pk"},
				EnvVars: []string{"WW_PRIVKEY"},
				Usage:   "path to private key file for libp2p identity",
			},
			&cli.PathFlag{
				Name:    "root",
				EnvVars: []string{"WW_ROOT"},
				Usage:   "ipfs path or local path to config directory",
				Value:   ".",
			},
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
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	// Initialize WebAssembly runtime with debug info if enabled
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).
		WithDebugInfoEnabled(c.Bool("wasm-debug")))
	defer r.Close(ctx)

	// Instantiate WASI preview1 for system interface compatibility
	wasi, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer wasi.Close(ctx)

	b, err := bytecode(c)
	if err != nil {
		return err
	}

	// Compile the WebAssembly module
	cm, err := r.CompileModule(ctx, b)
	if err != nil {
		return err
	}
	defer cm.Close(ctx)

	// Create a new process instance with:
	// - A unique process ID
	// - Command line arguments
	// - Standard error output
	// - A Unix-like filesystem
	p, err := proc.Command{
		PID:    proc.NewPID(),
		Args:   c.Args().Slice(),
		Env:    c.StringSlice("env"),
		Stderr: c.App.ErrWriter,
		FS:     env.NewUnixFS(ctx),
	}.Instantiate(ctx, r, cm)
	if err != nil {
		return err
	}
	defer p.Close(ctx)

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

	// Reserve resources for the bootloader execution
	if err := p.Reserve(ctx, struct {
		io.ReadCloser
		io.Writer
	}{
		ReadCloser: io.NopCloser(c.App.Reader),
		Writer:     c.App.Writer,
	}); err != nil {
		return err
	}
	defer p.Release()

	// Create router for inter-process message delivery
	rt := &Router{DB: db}

	// Initialize and bind system services to the supervisor:
	// - MDNS for local network discovery
	// - DHT for peer discovery over the network
	// - P2P for distributed communication
	// - HTTP API server
	for _, s := range []suture.Service{
		&glia.P2P{Env: &env, Router: rt},
		&glia.HTTP{
			Env:        &env,
			Root:       p.String(),
			Router:     rt,
			ListenAddr: c.String("http"),
		},
	} {
		wetware.Add(s)
	}
	cherr := wetware.ServeBackground(ctx)

	// Log successful server startup
	env.Log().Info("server started",
		"proc", p.String())

	// execute the bootloader
	if err := p.Method("_start").CallWithStack(ctx, nil); err != nil {
		if e, ok := err.(*sys.ExitError); !ok || e.ExitCode() != 0 {
			return err
		}
	}

	return <-cherr
}

func rootPath(c *cli.Context) (path.Path, error) {
	if c.NArg() == 0 {
		return path.NewPath(c.String("root"))
	}

	return path.NewPath(c.Args().First())
}

func bytecode(c *cli.Context) ([]byte, error) {
	root, err := rootPath(c)
	if err != nil {
		// Try loading from local file path first
		if _, err := os.Stat(c.Args().First()); err == nil {
			return os.ReadFile(c.Args().First())
		}
		return nil, err
	}

	return env.Load(c.Context, root)
}
