package ww

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/thejerf/suture/v4"
	"github.com/wetware/go/system"
	"github.com/wetware/go/vat"
)

const Proto = "/ww/0.0.0"

var _ suture.Service = (*Cluster)(nil)

type Config struct {
	NS     string
	IPFS   iface.CoreAPI
	Host   host.Host
	Router routing.Routing
}

func (cfg Config) Build(ctx context.Context) Cluster {
	p, err := cfg.IPFS.Name().Resolve(ctx, cfg.NS)
	if err != nil {
		defer slog.ErrorContext(ctx, "failed to build cluster",
			"ns", cfg.NS,
			"reason", err)
	}

	return Cluster{
		Err:    err,
		Root:   p,
		IPFS:   cfg.IPFS,
		Host:   cfg.Host,
		Router: cfg.Router,
	}
}

type Cluster struct {
	Err    error
	Root   path.Path
	IPFS   iface.CoreAPI
	Host   host.Host
	Router routing.Routing
}

func (c Cluster) String() string {
	peer := c.Host.ID()
	root := c.Root
	return fmt.Sprintf("%s::%s", peer, root)
}

func (c Cluster) Proto() protocol.ID {
	return protocol.ID(filepath.Join(Proto, c.Root.String()))
}

func (c Cluster) Setup(ctx context.Context) error {
	// build failed?
	if c.Err != nil {
		defer slog.Debug("skipped broken build",
			"path", c.Root,
			"error", c.Err)
		return suture.ErrDoNotRestart
	}

	if c.Router != nil {
		if err := c.Router.Bootstrap(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Serve the cluster's root process
func (c Cluster) Serve(ctx context.Context) error {
	if err := c.Setup(ctx); err != nil {
		return err
	}

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).
		WithDebugInfoEnabled(false))
	defer r.Close(ctx)

	cl, err := wasi.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer cl.Close(ctx)

	sys, err := system.Builder{
		// Host:    c.Host,
		// IPFS:    c.IPFS,
	}.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer sys.Close(ctx)

	cm, err := c.CompileModule(ctx, r)
	if err != nil {
		return err
	}
	defer cm.Close(ctx)

	mod, err := r.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
		// WithName().
		// WithArgs().
		// WithEnv().
		WithRandSource(rand.Reader).
		// WithFS().
		// WithFSConfig().
		// WithStartFunctions(). // remove _start so that we can call it later
		WithStdin(sys.Stdin()).
		WithStdout(os.Stdout). // FIXME
		WithStderr(os.Stderr). // FIXME
		WithSysNanotime())
	if err != nil {
		return err
	}
	defer mod.Close(ctx)

	net := vat.NetConfig{
		Host:   c.Host,
		Guest:  mod,
		System: sys,
	}.Build(ctx)
	defer net.Release()

	return net.Serve(ctx)
}

func (c Cluster) CompileModule(ctx context.Context, r wazero.Runtime) (wazero.CompiledModule, error) {
	f, err := c.ResolveRoot(ctx)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// FIXME:  address the obvious DoS vector
	bytecode, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return r.CompileModule(ctx, bytecode)
}

func (c Cluster) ResolveRoot(ctx context.Context) (files.File, error) {
	n, err := c.IPFS.Unixfs().Get(ctx, c.Root)
	if err != nil {
		return nil, err
	}

	return n.(files.File), nil
}
