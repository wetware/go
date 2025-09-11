//go:generate capnp compile -I.. -I$GOPATH/src/capnproto.org/go/capnp/std -ogo system.capnp

package system

import (
	"context"
	"crypto/rand"
	"io"
	"os"
	"runtime"

	capnp "capnproto.org/go/capnp/v3"
	server "capnproto.org/go/capnp/v3/server"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-msgio"
	"github.com/mr-tron/base58"
	"github.com/tetratelabs/wazero"
	"golang.org/x/sync/semaphore"
)

type TerminalConfig struct {
	Exec Executor
}

func (c TerminalConfig) New() Terminal {
	client := capnp.NewClient(c.NewServer())
	return Terminal(client)
}

func (c TerminalConfig) NewServer() *server.Server {
	return Terminal_NewServer(c)
}

func (c TerminalConfig) Login(ctx context.Context, call Terminal_login) error {
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	exec := c.Exec.AddRef()
	if err := results.SetExec(exec); err != nil {
		defer exec.Release()
		return err
	}

	return nil
}

type ExecutorConfig struct {
	Host    host.Host
	Runtime wazero.RuntimeConfig
}

func (c ExecutorConfig) New(ctx context.Context) Executor {
	client := capnp.NewClient(c.NewServer(ctx))
	return Executor(client)
}

func (c ExecutorConfig) NewServer(ctx context.Context) *server.Server {
	return Executor_NewServer(DefaultExecutor{
		Host:    c.Host,
		Runtime: wazero.NewRuntimeWithConfig(ctx, c.Runtime),
	})
}

type DefaultExecutor struct {
	Background context.Context
	Runtime    wazero.Runtime
	Host       host.Host
}

func (d DefaultExecutor) Exec(ctx context.Context, call Executor_exec) error {
	b, err := call.Args().Bytecode()
	if err != nil {
		return err
	}

	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	var ok = false
	cm, err := d.Runtime.CompileModule(ctx, b)
	if err != nil {
		return err
	}
	defer func() {
		if !ok {
			cm.Close(ctx)
		}
	}()

	e := NewEndpoint()
	mod, err := d.Runtime.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
		WithName(e.String()).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithStdin(e).
		WithStdout(e).
		WithStderr(os.Stderr))
	if err != nil {
		return err
	}
	defer func() {
		if !ok {
			e.Close()
		}
	}()

	d.Host.SetStreamHandler(e.Protocol(), func(s network.Stream) {
		defer s.Close()

		r := msgio.NewVarintReader(s)
		defer r.Close()

		b, err := r.ReadMsg()
		if err != nil {
			return
		}
		defer r.ReleaseMsg(b)

		m, err := capnp.Unmarshal(b)
		if err != nil {
			return
		}
		defer m.Release()

		call, err := ReadRootMethodCall(m)
		if err != nil {
			return
		}

		name, err := call.Name()
		if err != nil {
			return
		}
		method := mod.ExportedFunction(name)
		if method == nil {
			return
		}

		var stack []uint64
		if params, err := call.Stack(); err == nil && params.Len() > 0 {
			n := params.Len()
			if np := len(method.Definition().ParamTypes()); np > n {
				n = np
			}
			if nr := len(method.Definition().ResultTypes()); nr > n {
				n = nr
			}
			if n > 0 { // only allocate if len(stack) is going to be nonzero
				stack = make([]uint64, n)
				for i := 0; i < params.Len(); i++ {
					stack[i] = params.At(i)
				}
			}
		}

		if err := e.sem.Acquire(ctx, 1); err != nil {
			return
		}
		defer e.sem.Release(1)

		e.ReadWriteCloser = s
		defer func() {
			e.ReadWriteCloser = nil
		}()

		if err := method.CallWithStack(ctx, stack); err != nil {
			return
		}
	})
	defer func() {
		if !ok {
			d.Host.RemoveStreamHandler(e.Protocol())
		}
	}()

	if err = res.SetProtocol(e.String()); err == nil {
		// We succeeded in writing our process' protocol ID to the response,
		// so now we want our WASM objects to outlive the scope of this function.
		ok = true
		runtime.SetFinalizer(e, func(e *Endpoint) {
			// Endpoint is released after stream handler is removed.
			mod.Close(ctx)
			cm.Close(ctx)
		})
	}
	return err
}

func newName() string {
	var buf [8]byte
	if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
		panic(err)
	}
	return base58.FastBase58Encoding(buf[:])
}

type Endpoint struct {
	Name string
	io.ReadWriteCloser
	sem *semaphore.Weighted
}

func NewEndpoint() *Endpoint {
	return &Endpoint{
		Name: newName(),
		sem:  semaphore.NewWeighted(1),
	}
}

func (e Endpoint) String() string {
	proto := e.Protocol()
	return string(proto)
}

func (e Endpoint) Protocol() protocol.ID {
	return protocol.ID("/ww/0.1.0/" + e.Name)
}
