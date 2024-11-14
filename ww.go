package ww

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/blang/semver/v4"
	"github.com/libp2p/go-libp2p/core/event"
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

	return env.ServeProc(ctx, &p)
}

func (env Env) ServeProc(ctx context.Context, p *proc.P) error {
	sub, err := env.Host.EventBus().Subscribe([]any{
		new(event.EvtLocalAddressesUpdated),
		new(event.EvtLocalProtocolsUpdated)})
	if err != nil {
		return err
	}
	defer sub.Close()

	// if err := env.Bootstrap(ctx, &p); err != nil {
	// 	return err
	// }

	// TODO:  client apps shouldn't listen for streams
	release := env.Bind(ctx, p)
	defer release()

	slog.InfoContext(ctx, "event loop started",
		"peer", env.Host.ID(),
		"proc", p.String())
	defer slog.WarnContext(ctx, "event loop halted",
		"peer", env.Host.ID(),
		"proc", p.String())

	for {
		var v any
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v = <-sub.Out(): // assign event to v
		}

		switch ev := v.(type) {
		case *event.EvtLocalAddressesUpdated:
			slog.InfoContext(ctx, "local addresses updated",
				"peer", env.Host.ID(),
				"current", ev.Current,
				"removed", ev.Removed,
				"diffs", ev.Diffs)

		case *event.EvtLocalProtocolsUpdated:
			slog.InfoContext(ctx, "local protocols updated",
				"peer", env.Host.ID(),
				"added", ev.Added,
				"removed", ev.Removed)
		}
	}
}

func (env Env) Bootstrap(ctx context.Context, p *proc.P) error {
	b, err := io.ReadAll(&io.LimitedReader{
		R: env.Stdin,
		N: int64(1<<32 - 1), // max u32
	})
	if err != nil || len(b) == 0 {
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

func (env Env) Bind(ctx context.Context, p *proc.P) ReleaseFunc {
	var timeout time.Duration // default 0
	if dl, ok := ctx.Deadline(); ok {
		timeout = time.Until(dl)
	}

	handler := proc.StreamHandler{
		Proc:               p,
		MessageReadTimeout: timeout}
	proto := handler.Proto()

	env.Host.SetStreamHandlerMatch(
		proto,
		handler.Match,
		handler.Bind(ctx))
	return func() { env.Host.RemoveStreamHandler(proto) }
}

type ReleaseFunc func()
