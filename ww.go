package ww

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ipfs/boxo/files"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
	"github.com/thejerf/suture/v4"
)

const Proto = "/ww/0.0.0"

var _ suture.Service = (*Cluster)(nil)

type Cluster struct {
	Host host.Host
	IPFS iface.CoreAPI
	Name string
}

func (c Cluster) String() string {
	return fmt.Sprintf("Cluster{cid=%s}", c.Name)
}

func (c Cluster) Serve(ctx context.Context) error {
	n, err := c.ResolveRoot(ctx)
	if err != nil {
		return err
	}
	defer n.Close()

	// proto := protocol.ID(path.Join(Proto, p.String()))
	// c.Host.SetStreamHandlerMatch(proto, c.Match(n), c.NewHandler(ctx, n))
	// defer c.Host.RemoveStreamHandler(proto)

	// Run root process
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).
		WithDebugInfoEnabled(false))
	defer r.Close(ctx)

	cl, err := wasi.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer cl.Close(ctx)

	// FIXME:  address the obvious DoS vector
	bytecode, err := io.ReadAll(n.(files.File))
	if err != nil {
		return err
	}

	cm, err := r.CompileModule(ctx, bytecode)
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
		WithStartFunctions().
		// WithStdin().
		WithStdout(os.Stdout). // FIXME
		WithStderr(os.Stderr). // FIXME
		WithSysNanotime())
	if err != nil {
		return err
	}
	defer mod.Close(ctx)

	start := mod.ExportedFunction("_start")
	if start == nil {
		return errors.New("exported function not found: _start")
	}

	_, err = start.Call(ctx)
	if e, ok := err.(*sys.ExitError); ok && e.ExitCode() != 0 {
		return err
	}

	return suture.ErrDoNotRestart
}

func (c Cluster) ResolveRoot(ctx context.Context) (files.Node, error) {
	p, err := c.IPFS.Name().Resolve(ctx, c.Name)
	if err != nil {
		return nil, err
	}

	return c.IPFS.Unixfs().Get(ctx, p)
}

// func (c Cluster) Match(n ipld.Node) func(protocol.ID) bool {
// 	return func(id protocol.ID) bool {
// 		return strings.HasPrefix(string(id), "/"+n.String())
// 	}
// }

// func (c Cluster) NewHandler(ctx context.Context, n ipld.Node) func(network.Stream) {
// 	return func(s network.Stream) {
// 		defer s.Close()

// 		path := string(s.Protocol())
// 		conn := rpc.NewConn(rpc.NewPackedStreamTransport(s), &rpc.Options{
// 			BootstrapClient: c.ResolveBootstrap(ctx, n, path),
// 		})
// 		defer conn.Close()

// 		select {
// 		case <-ctx.Done():
// 		case <-conn.Done():
// 		}
// 	}
// }

// func (c Cluster) ResolveBootstrap(ctx context.Context, n ipld.Node, path string) capnp.Client {
// 	client, resolver := capnp.NewLocalPromise[capnp.Client]()

// 	go func() {
// 		var v any
// 		var err error
// 		var p = c.ParsePath(path)
// 		for {
// 			v, p, err = n.Resolve(p)
// 			if err != nil {
// 				resolver.Reject(err)
// 				return
// 			}

// 			switch x := v.(type) {
// 			case ipld.Link:
// 				if n, err = c.IPFS.Dag().Get(ctx, x.Cid); err != nil {
// 					resolver.Reject(err)
// 					return
// 				}

// 			case ipld.Node:
// 				resolver.Fulfill(capnp.Client{}) // FIXME
// 				// resolver.Fulfill(makeClientFromNode(x)) // TODO:  load WASM from node, compile & run.
// 				return
// 			}
// 		}
// 	}()

// 	return client
// }

// func (c Cluster) ParsePath(s string) []string {
// 	p := path.Clean(s)
// 	p = strings.TrimLeft(p, ".")
// 	p = strings.TrimPrefix(p, Proto)
// 	p = strings.TrimPrefix(p, c.Name)
// 	return strings.Split(p, "/")
// }
