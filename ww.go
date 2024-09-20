package ww

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"runtime"

	"capnproto.org/go/capnp/v3"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/thejerf/suture/v4"
	"github.com/wetware/go/auth"
	"github.com/wetware/go/system"
	"go.uber.org/multierr"
)

var _ suture.Service = (*Cluster)(nil)

type UnixFS interface {
	fs.FS
	Resolve(context.Context, string) (path.Path, files.Node, error)
}

type Config struct {
	NS     string
	Host   host.Host
	UnixFS UnixFS
	Cache  wazero.CompilationCache
	Stdio  struct {
		Reader        io.Reader
		Writer, Error io.WriteCloser
	}
}

func (config Config) Build() Cluster {
	// HACK:  right now we just have one capability
	// and it's stdio, so always allow it.  This
	// effectively disables auth.  We'll get back
	// to it soon.
	baseCaps := provide{
		Stdio: stdio{
			Reader: config.Stdio.Reader,
			Writer: config.Stdio.Writer,
			Error:  config.Stdio.Error,
		},
	}

	return Cluster{
		Config: config,
		System: auth.TerminalConfig{
			Rand: rand.Reader,
			Auth: baseCaps,
		}.Build(),
	}
}

func (config Config) CompilationCache() wazero.CompilationCache {
	if config.Cache == nil {
		config.Cache = wazero.NewCompilationCache()
	}

	return config.Cache
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
		WithCompilationCache(c.Config.CompilationCache()).
		WithCloseOnContextDone(true))
	defer r.Close(ctx)

	wasi, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer wasi.Close(ctx)

	compiled, err := c.Compile(ctx, r)
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
		WithFS(c.UnixFS))
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

	return auth.Session{
		Proc: f.Stdio(),
		Sock: system.Socket{},
	}, release
}

func (c Cluster) Compile(ctx context.Context, r wazero.Runtime) (wazero.CompiledModule, error) {
	_, n, err := c.UnixFS.Resolve(ctx, ".")
	if err != nil {
		return nil, err
	}
	defer n.Close()

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

type provide struct {
	Stdio interface {
		BindReader() auth.ReadPipe
		BindWriter() auth.WritePipe
		BindError() auth.WritePipe
	}
}

func (p provide) BindPolicy(user crypto.PubKey, policy auth.Policy) error {
	sock, err := policy.NewStdio()
	if err == nil {
		err = p.BindStdio(sock)
	}

	return err
}

func (p provide) BindStdio(sock auth.Socket) error {
	return multierr.Combine(
		sock.SetReader(p.Stdio.BindReader()),
		sock.SetWriter(p.Stdio.BindWriter()),
		sock.SetError(p.Stdio.BindError()))
}

type stdio struct {
	Reader io.Reader
	Writer io.WriteCloser
	Error  io.WriteCloser
}

func (s stdio) BindReader() auth.ReadPipe {
	return system.NewReadPipe(s.Reader)
}

func (s stdio) BindWriter() auth.WritePipe {
	return system.NewWritePipe(s.Writer)
}

func (s stdio) BindError() auth.WritePipe {
	return system.NewWritePipe(s.Error)
}
