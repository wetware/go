package ww

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/thejerf/suture/v4"
	"github.com/wetware/go/system"
	"github.com/wetware/go/vat"
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
}

func (cfg Config) Build(ctx context.Context) Cluster {
	return Cluster{
		Config: cfg,
	}
}

type Cluster struct {
	Config
}

func (c Cluster) String() string {
	peer := c.Host.ID()
	return fmt.Sprintf("Cluster{peer=%s}", peer)
}

func (c Cluster) Bootstrap(ctx context.Context) error {
	if c.Router == nil {
		slog.WarnContext(ctx, "no router",
			"cluster", c.NS)
		return nil
	}

	return c.Router.Bootstrap(ctx)
}

// Serve the cluster's root process
func (c Cluster) Serve(ctx context.Context) error {
	if err := c.Bootstrap(ctx); err != nil {
		return err
	}

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithMemoryLimitPages(1024). // 64MB
		WithCloseOnContextDone(true).
		WithDebugInfoEnabled(c.Debug))
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
	bytecode, err := c.LoadByteCode(ctx)
	if err != nil {
		return nil, err
	}

	return r.CompileModule(ctx, bytecode)
}

func (c Cluster) LoadByteCode(ctx context.Context) ([]byte, error) {
	root, err := c.IPFS.Name().Resolve(ctx, c.NS)
	if err != nil {
		return nil, err
	}

	n, err := c.IPFS.Unixfs().Get(ctx, root)
	if err != nil {
		return nil, err
	}
	defer n.Close()

	// FIXME:  address the obvious DoS vector
	return io.ReadAll(n.(files.File))
}
