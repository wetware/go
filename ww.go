package ww

import (
	"context"
	"fmt"
	"log/slog"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/thejerf/suture/v4"
	"github.com/wetware/go/guest"
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

// Serve the cluster's root process
func (c Cluster) Serve(ctx context.Context) error {
	if c.Router == nil {
		slog.WarnContext(ctx, "started with null router",
			"ns", c.NS)
		return nil
	}

	if err := c.Router.Bootstrap(ctx); err != nil {
		return err
	}

	root, err := c.IPFS.Name().Resolve(ctx, c.NS)
	if err != nil {
		return err
	}

	return c.ServeVat(ctx, root)
}

func (c Cluster) ServeVat(ctx context.Context, root path.Path) error {
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

	mod, err := guest.Config{
		IPFS: c.IPFS,
		Root: root,
		Sys:  sys,
	}.Instanatiate(ctx, r)
	if err != nil {
		return err
	}
	defer mod.Close(ctx)

	// Obtain the system client.  This gives us an API to our root
	// process.
	client := sys.Boot(mod)
	defer client.Release()

	net := vat.Config{
		Host:  c.Host,
		Proto: vat.ProtoFromModule(mod),
	}.Build(ctx)
	defer net.Release()

	for {
		if conn, err := net.Accept(ctx, &rpc.Options{
			BootstrapClient: client.AddRef(),
			Network:         net,
		}); err == nil {
			go net.ServeConn(ctx, conn)
		} else {
			return err
		}
	}
}
