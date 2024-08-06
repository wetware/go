package ww

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/thejerf/suture/v4"
	"github.com/wetware/go/system"
)

const Proto = "/ww/0.0.0"

var _ suture.Service = (*Cluster)(nil)

type Config struct {
	NS      string
	IPFS    iface.CoreAPI
	Host    host.Host
	IO      system.Streams
	Runtime wazero.RuntimeConfig
}

func (config Config) Build() Cluster {
	if config.Runtime == nil {
		config.Runtime = wazero.NewRuntimeConfig().
			// WithCompilationCache().
			WithCloseOnContextDone(true)
	}

	// Use the public IPFS DHT for routing.
	config.Host = routedhost.Wrap(
		config.Host,
		config.IPFS.Routing())

	return Cluster{
		Config: config,
	}
}

type Cluster struct {
	Config
}

func (c Cluster) String() string {
	return c.Config.NS
}

// Serve the cluster's root filesystem
func (c Cluster) Serve(ctx context.Context) error {
	fs, err := c.NewFS(ctx)
	if err != nil {
		return err
	}

	root, err := c.Resolve(ctx, fs.Root)
	if err != nil {
		return err
	}
	defer root.Close()

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(ctx)

	wasi, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer wasi.Close(ctx)

	compiled, err := c.CompileNode(ctx, r, root)
	if err != nil {
		return err
	}
	defer compiled.Close(ctx)

	mod, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().
		WithStartFunctions(). // don't call _start automatically
		WithName(fs.Root.String()).
		// WithArgs().
		// WithEnv().
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithRandSource(rand.Reader).
		WithOsyield(runtime.Gosched).
		WithStdin(c.IO.Stdin()).
		WithStdout(c.IO.Stdout()).
		WithStderr(c.IO.Stderr()).
		WithFS(fs))
	if err != nil {
		return err
	}
	defer mod.Close(ctx)

	_, err = mod.ExportedFunction("_start").Call(ctx)
	return err
}

// NewFS returns an fs.FS.
func (c Cluster) NewFS(ctx context.Context) (*system.FS, error) {
	root, err := path.NewPath(c.NS)
	if err != nil {
		return nil, err
	}

	return &system.FS{
		Ctx:  ctx,
		API:  c.IPFS.Unixfs(),
		Root: root,
	}, nil
}

// Resolve an IPFS path into a virtual filesystem node.
func (c Cluster) Resolve(ctx context.Context, p path.Path) (n files.Node, err error) {
	switch ns := p.Segments()[0]; ns {
	case "ipns":
		// IPNS introduces one level of indirection:  a mutable name.
		// Here we are fetching the IPFS record to which the name is
		// currently pointing.
		p, err = c.IPFS.Name().Resolve(ctx, p.String())
		if err != nil {
			return
		}

	default:
		slog.Debug("resolved namespace",
			"ns", ns)
	}

	n, err = c.IPFS.Unixfs().Get(ctx, p)
	return
}

// CompileNode reads bytecode from an IPFS node and compiles it.
func (c Cluster) CompileNode(ctx context.Context, r wazero.Runtime, node files.Node) (wazero.CompiledModule, error) {
	bytecode, err := c.LoadByteCode(ctx, node)
	if err != nil {
		return nil, err
	}

	return r.CompileModule(ctx, bytecode)
}

// LoadByteCode from the provided IPFS node.
func (c Cluster) LoadByteCode(ctx context.Context, node files.Node) (b []byte, err error) {
	err = files.Walk(node, func(fpath string, node files.Node) error {
		if b != nil {
			return errAbortWalk
		}

		switch fname := filepath.Base(fpath); fname {
		case "main.wasm":
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
