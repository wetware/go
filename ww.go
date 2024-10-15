package ww

import (
	"context"
	"errors"

	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

const Version = "0.1.0"
const Proto = "/ww/" + Version

type ReleaseFunc func()

type Loader interface {
	Load(ctx context.Context) ([]byte, error)
}

type Env struct {
	IPFS   iface.CoreAPI
	Host   host.Host
	Boot   Loader
	WASM   wazero.RuntimeConfig
	Module wazero.ModuleConfig
}

func (env Env) Serve(ctx context.Context) error {
	bytecode, err := env.Boot.Load(ctx)
	if err != nil {
		return err
	}

	return env.CompileAndServe(ctx, bytecode)
}

// Compile and serve the supplied bytecode using the supplied runtime.
func (env Env) CompileAndServe(ctx context.Context, bytecode []byte) error {
	r := wazero.NewRuntimeWithConfig(ctx, env.WASM)
	defer r.CloseWithExitCode(ctx, 0)

	wasi, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer wasi.Close(ctx)

	cm, err := r.CompileModule(ctx, bytecode)
	if err != nil {
		return err
	}
	defer cm.Close(ctx)

	mod, err := r.InstantiateModule(ctx, cm, env.Module)
	if err != nil {
		return err
	}
	defer mod.Close(ctx)

	return env.ServeProc(ctx, mod)
}

// ServeProc starts the process and blocks until it terminates or the
// context is canceled.
func (env Env) ServeProc(ctx context.Context, mod api.Module) error {
	fn := mod.ExportedFunction("_start")
	if fn == nil {
		return errors.New("missing export: _start")
	}

	_, err := fn.Call(ctx)
	switch e := err.(type) {
	case interface{ ExitCode() uint32 }:
		switch e.ExitCode() {
		case 0:
			err = nil
		case sys.ExitCodeContextCanceled:
			err = context.Canceled
		case sys.ExitCodeDeadlineExceeded:
			err = context.DeadlineExceeded
		}
	}

	return err
}

// func (env Env) BindSocket(ctx context.Context, mod api.Module, sock io.Writer) ReleaseFunc {
// 	// Semaphore with max weight of 1 acts as an unbounded mpmc queue.
// 	sem := semaphore.NewWeighted(1)

// 	// Handler is called in its own goroutine for each incoming libp2p stream.
// 	handler := func(s network.Stream) {
// 		defer s.Close()

// 		// did we get the lock?
// 		if sem.Acquire(ctx, 1) == nil {
// 			defer sem.Release(1)

// 			if n, err := io.Copy(sock, s); err != nil {
// 				slog.Warn("message delivery failed",
// 					"bytes_written", n,
// 					"reason", err)
// 			}
// 		}
// 	}

// 	proto := env.Proto(mod)
// 	matchPrefix := func(id protocol.ID) bool {
// 		return strings.HasPrefix(string(id), string(proto))
// 	}

// 	env.Host.SetStreamHandlerMatch(proto, matchPrefix, handler)
// 	return func() { env.Host.RemoveStreamHandler(proto) }
// }

// func (env Env) Proto(mod interface{ Name() string }) protocol.ID {
// 	id := filepath.Join(Proto, mod.Name())
// 	return protocol.ID(id)
// }
