package ww

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/thejerf/suture/v4"
	"github.com/wetware/go/system"
)

const Proto = "/ww/0.0.0"

var _ suture.Service = (*Cluster)(nil)

type Resolver interface {
	Resolve(ctx context.Context, ns string) (path.Path, error)
}

type Config struct {
	NS     string
	IPFS   iface.CoreAPI
	Host   host.Host
	Router routing.Routing
	Debug  bool // debug info enabled
	Stdio  struct {
		Reader    io.Reader
		Writer    io.WriteCloser
		ErrWriter io.WriteCloser
	}
}

func (cfg Config) Build(ctx context.Context) Cluster {
	return Cluster{
		Config: cfg,
	}
}

func (c Config) Stdin() io.Reader {
	if c.Stdio.Reader == nil {
		return os.Stdin
	}

	return c.Stdio.Reader
}

func (c Config) Stdout() io.WriteCloser {
	if c.Stdio.Writer == nil {
		return os.Stdout
	}

	return c.Stdio.Writer
}

func (c Config) Stderr() io.WriteCloser {
	if c.Stdio.ErrWriter == nil {
		return os.Stdout
	}

	return c.Stdio.ErrWriter
}

type Cluster struct {
	Config
}

func (c Cluster) String() string {
	return c.Config.NS
}

// Serve the cluster's root process
func (c Cluster) Serve(ctx context.Context) error {
	root, err := path.NewPath(c.NS)
	if err != nil {
		return err
	}

	node, err := c.Resolve(ctx, root)
	if err != nil {
		return err
	}
	defer node.Close()

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(ctx)

	wasi, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer wasi.Close(ctx)

	bytecode, err := c.LoadByteCode(ctx, node)
	if err != nil {
		return err
	}

	compiled, err := r.CompileModule(ctx, bytecode)
	if err != nil {
		return err
	}
	defer compiled.Close(ctx)

	mod, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().
		WithStartFunctions(). // don't call _start automatically
		WithName(c.NS).
		// WithArgs().
		// WithEnv().
		WithStdin(c.Stdin()).
		WithStderr(c.Stderr()).
		WithStdout(c.Stdout()).
		WithFS(system.FS{Ctx: ctx, API: c.IPFS.Unixfs(), Root: root}).
		WithRandSource(rand.Reader).
		WithOsyield(runtime.Gosched))
	if err != nil {
		return err
	}
	defer mod.Close(ctx)

	_, err = mod.ExportedFunction("_start").Call(ctx)
	return err
}

func (c Cluster) Resolve(ctx context.Context, root path.Path) (n files.Node, err error) {
	switch ns := root.Segments()[0]; ns {
	case "ipld":
		// IPLD introduces one level of indirection:  a mutable name.
		// Here we are fetching the IPFS record to which the name is
		// currently pointing.
		root, err = c.IPFS.Name().Resolve(ctx, root.String())
		if err != nil {
			return
		}

	default: // It's probably /ipfs/
	}

	n, err = c.IPFS.Unixfs().Get(ctx, root)
	return
}

func (c Cluster) LoadByteCode(ctx context.Context, node files.Node) (b []byte, err error) {
	err = files.Walk(node, func(fpath string, node files.Node) error {
		if b != nil {
			return errAbortWalk
		}

		switch fname := filepath.Base(fpath); fname {
		case "main.wasm":
			slog.InfoContext(ctx, "loading file",
				"name", fname,
				"path", fpath)

			if f := files.ToFile(node); f != nil {
				b, err = io.ReadAll(f)
				return err
			}
		}

		return nil
	})
	if err == errAbortWalk {
		err = nil
	}

	return
}

var errAbortWalk = errors.New("abort walk")
