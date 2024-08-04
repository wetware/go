package ww

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"

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
	return c.Config.NS
}

// Serve the cluster's root process
func (c Cluster) Serve(ctx context.Context) error {
	root, err := c.IPFS.Name().Resolve(ctx, c.NS)
	if err != nil {
		return err
	}

	node, err := c.IPFS.Unixfs().Get(ctx, root)
	if err != nil {
		return err
	}
	defer node.Close()

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(ctx)

	switch n := node.(type) {
	case files.File:
		// assume it's a WASM file; run it
		body := io.LimitReader(n, 2<<32)
		b, err := io.ReadAll(body)
		if err != nil {
			return err
		}

		r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
			WithCloseOnContextDone(true))
		defer r.Close(ctx)

		cl, err := wasi_snapshot_preview1.Instantiate(ctx, r)
		if err != nil {
			return err
		}
		defer cl.Close(ctx)

		cm, err := r.CompileModule(ctx, b)
		if err != nil {
			return err
		}
		defer cm.Close(ctx)

		mod, err := r.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
			// WithArgs().
			// WithEnv().
			// WithNanosleep().
			// WithNanotime().
			// WithOsyield().
			// WithSysNanosleep().
			// WithSysNanotime().
			// WithSysWalltime().
			// WithWalltime().
			WithStartFunctions().
			WithFS(system.FS{API: c.IPFS.Unixfs()}).
			WithRandSource(rand.Reader).
			WithStdin(os.Stdin).
			WithStderr(os.Stderr).
			WithStdout(os.Stdout).
			WithName(c.NS))
		if err != nil {
			return err
		}
		defer mod.Close(ctx)

		_, err = mod.ExportedFunction("_start").Call(ctx)
		return err

	case files.Directory:
		// TODO:  look for a main.wasm and execute it
		return errors.New("Cluster.Serve::TODO:implement directory handler")

	default:
		return fmt.Errorf("unhandled type: %s", reflect.TypeOf(n))
	}
}
