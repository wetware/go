package ww

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/thejerf/suture/v4"
)

const Proto = "/ww/0.0.0"

var _ suture.Service = (*Cluster)(nil)

type Cluster struct {
	Root path.Path
	IPFS iface.CoreAPI
	Host host.Host
}

func (c Cluster) String() string {
	return fmt.Sprintf("Cluster{%s}", c.Root)
}

func (c Cluster) Proto() protocol.ID {
	return protocol.ID(filepath.Join(Proto, c.Root.String()))
}

// Serve the cluster's root processs
func (c Cluster) Serve(ctx context.Context) error {
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).
		WithDebugInfoEnabled(false))
	defer r.Close(ctx)

	cl, err := wasi.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer cl.Close(ctx)

	// sys, err := system.Builder{
	// 	// Host:    c.Host,
	// 	// IPFS:    c.IPFS,
	// 	Runtime: r,
	// }.Instantiate(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer sys.Close(ctx)

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
		// WithStdin().
		WithStdout(os.Stdout). // FIXME
		WithStderr(os.Stderr). // FIXME
		WithSysNanotime())
	if err != nil {
		return err
	}
	defer mod.Close(ctx)

	<-ctx.Done()
	return ctx.Err()
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
