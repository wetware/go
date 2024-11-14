package ww

import (
	"context"
	"io"
	"io/fs"

	"capnproto.org/go/capnp/v3"
	"github.com/blang/semver/v4"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/wetware/go/proc"
	protoutils "github.com/wetware/go/util/proto"
)

const Version = "0.1.0"

var Proto = protoutils.VersionedID{
	ID:      "ww",
	Version: semver.MustParse(Version),
}

type Env struct {
	Args    []string
	Vars    []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Host    host.Host
	Runtime wazero.Runtime
	Root    string
	FS      fs.FS
}

func (env Env) Serve(ctx context.Context) error {
	f, err := env.FS.Open(env.Root)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	return env.CompileAndServe(ctx, b)
}

// Compile and serve the supplied bytecode using the supplied runtime.
func (env Env) CompileAndServe(ctx context.Context, bytecode []byte) error {
	wasi, err := wasi_snapshot_preview1.Instantiate(ctx, env.Runtime)
	if err != nil {
		return err
	}
	defer wasi.Close(ctx)

	cm, err := env.Runtime.CompileModule(ctx, bytecode)
	if err != nil {
		return err
	}
	defer cm.Close(ctx)

	// Instantiate and bind the root process.
	var p proc.P
	err = proc.Config{
		Proto:   Proto.Unwrap(),
		Args:    env.Args,
		Env:     env.Vars,
		Stdout:  env.Stdout,
		Stderr:  env.Stderr,
		Runtime: env.Runtime,
		Module:  cm,
	}.Bind(ctx, &p)
	if err != nil {
		return err
	}
	defer p.Close(ctx)

	if err := env.Bootstrap(ctx, &p); err != nil {
		return err
	}

	release := Bind(ctx, env.Host, &p)
	defer release()

	<-ctx.Done()
	return ctx.Err()
}

func (env Env) Bootstrap(ctx context.Context, p *proc.P) error {
	b, err := io.ReadAll(&io.LimitedReader{
		R: env.Stdin,
		N: int64(1<<32 - 1), // max u32
	})
	if err != nil {
		return err
	}

	m, err := capnp.Unmarshal(b)
	if err != nil {
		return err
	}
	defer m.Release()

	call, err := proc.ReadRootMethodCall(m)
	if err != nil {
		return err
	}

	return p.Deliver(ctx, call)
}

type ReleaseFunc func()

func Bind(ctx context.Context, h host.Host, p *proc.P) ReleaseFunc {
	handler := proc.StreamHandler{Proc: p}
	proto := handler.Proto()

	h.SetStreamHandlerMatch(
		proto,
		handler.Match,
		handler.Bind(ctx))
	return func() { h.RemoveStreamHandler(proto) }
}
