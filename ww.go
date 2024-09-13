package ww

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"

	"capnproto.org/go/capnp/v3"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/thejerf/suture/v4"
	"github.com/wetware/go/auth"
	"github.com/wetware/go/system"
)

const Proto = "/ww/0.0.0"

var _ suture.Service = (*Cluster)(nil)

type Config struct {
	NS   string
	IPFS iface.CoreAPI
	Host host.Host
	// IO      system.Streams
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
		System: auth.TerminalConfig{
			Rand: rand.Reader,
			Auth: auth.Provide{
				// ...
			},
		}.Build(),
	}
}

type Cluster struct {
	Config
	System auth.Terminal
}

func (c Cluster) String() string {
	return c.Config.NS
}

// Serve the cluster's root filesystem
func (c Cluster) Serve(ctx context.Context) error {
	sess, release := c.Login(ctx)
	defer release()

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(ctx)

	wasi, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer wasi.Close(ctx)

	sys, err := system.HostConfig{
		NS:   c.NS,
		Host: c.Host,
	}.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer sys.Close(ctx)

	fs, err := c.NewFS(ctx)
	if err != nil {
		return err
	}

	compiled, err := c.Compile(ctx, r, fs)
	if err != nil {
		return err
	}
	defer compiled.Close(ctx)

	mod, err := r.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().
		WithStartFunctions(). // don't call _start automatically
		WithName(compiled.Name()).
		// WithArgs().
		// WithEnv().
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithRandSource(rand.Reader).
		WithOsyield(runtime.Gosched).
		WithStdin(sess.Reader(ctx)).
		WithStdout(sess.Writer(ctx)).
		WithStderr(sess.ErrWriter(ctx)).
		WithFS(fs))
	if err != nil {
		return err
	}
	defer mod.Close(ctx)

	_, err = mod.ExportedFunction("_start").Call(ctx)
	return err
}

func (c Cluster) Login(ctx context.Context) (auth.Session, capnp.ReleaseFunc) {
	pk := c.Host.Peerstore().PrivKey(c.Host.ID())
	signer := &auth.SignOnce{PrivKey: pk}
	account := auth.Signer_ServerToClient(signer)

	f, release := c.System.Login(ctx, func(t auth.Terminal_login_Params) error {
		return t.SetAccount(account)
	})

	f.Stdio().Reader()

	return auth.Session{
		Proc: f.Stdio(),
		Sock: system.Socket{},
	}, release
}

// NewFS returns an fs.FS.
func (c Cluster) NewFS(ctx context.Context) (*system.IPFS, error) {
	root, err := path.NewPath(c.NS)
	if err != nil {
		return nil, err
	}

	return &system.IPFS{
		Ctx:  ctx,
		Unix: c.IPFS.Unixfs(),
		Root: root,
	}, nil
}

func (c Cluster) Compile(ctx context.Context, r wazero.Runtime, fs *system.IPFS) (wazero.CompiledModule, error) {
	root, n, err := fs.Resolve(ctx, ".")
	if err != nil {
		return nil, err
	}
	defer n.Close()

	slog.DebugContext(ctx, "resolved root",
		"path", root)

	return c.CompileNode(ctx, r, n)
}

// CompileNode reads bytecode from an IPFS node and compiles it.
func (c Cluster) CompileNode(ctx context.Context, r wazero.Runtime, node files.Node) (wazero.CompiledModule, error) {
	bytecode, err := c.LoadByteCode(ctx, node)
	if err != nil {
		return nil, err
	}

	return r.CompileModule(ctx, bytecode)
}

// LoadByteCode loads the bytecode from the provided IPFS node.
// If the node is a directory, it will walk the directory and
// load the bytecode from the first file named "main.wasm". If
// the node is a file, it will attempt to load the bytecode from
// the file.  An error from Wazero usually indicates that the
// bytecode is invalid.
func (c Cluster) LoadByteCode(ctx context.Context, node files.Node) ([]byte, error) {
	switch node := node.(type) {
	case files.File:
		return io.ReadAll(node)

	case files.Directory:
		return c.LoadByteCodeFromDir(ctx, node)

	default:
		panic(node) // unreachable
	}
}

func (c Cluster) LoadByteCodeFromDir(ctx context.Context, d files.Directory) (b []byte, err error) {
	if err = files.Walk(d, func(fpath string, node files.Node) error {
		// Note:  early returns are used to short-circuit the walk. These
		// are signaled by returning errAbortWalk.

		// Already have the bytecode?
		if b != nil {
			return errAbortWalk
		}

		// File named "main.wasm"?
		if fname := filepath.Base(fpath); fname == "main.wasm" {
			if b, err = c.LoadByteCode(ctx, node); err != nil {
				return err
			}

			return errAbortWalk
		}

		// Keep walking.
		return nil
	}); err == errAbortWalk { // no error; we've just bottomed out
		err = nil
	}

	return
}

var errAbortWalk = errors.New("abort walk")
