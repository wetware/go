//go:generate capnp compile -I.. -I$GOPATH/src/capnproto.org/go/capnp/std -ogo system.capnp

package system

import (
	"context"
	"crypto/rand"
	"io"
	"os"

	capnp "capnproto.org/go/capnp/v3"
	server "capnproto.org/go/capnp/v3/server"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
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

	proc, err := ProcConfig{
		Host:      d.Host,
		Runtime:   d.Runtime,
		Bytecode:  b,
		ErrWriter: os.Stderr, // HACK
	}.New(ctx)
	if err != nil {
		return err
	} else if err = res.SetProtocol(proc.String()); err != nil {
		defer proc.Close(ctx)
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

func (e *Endpoint) Close() error {
	if e.ReadWriteCloser != nil {
		return e.ReadWriteCloser.Close()
	}
	return nil
}
