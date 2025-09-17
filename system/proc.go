package system

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"go.uber.org/multierr"
)

type ProcConfig struct {
	Host      host.Host
	Runtime   wazero.Runtime
	Bytecode  []byte
	ErrWriter io.Writer
}

func (c ProcConfig) New(ctx context.Context) (*Proc, error) {
	var ok = false
	cm, err := c.Runtime.CompileModule(ctx, c.Bytecode)
	if err != nil {
		return nil, err
	}

	sys, err := wasi_snapshot_preview1.Instantiate(ctx, c.Runtime)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !ok {
			sys.Close(ctx)
		}
	}()

	e := NewEndpoint()
	mod, err := c.Runtime.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
		WithName(e.String()).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithStdin(e).
		WithStdout(e).
		WithStderr(c.ErrWriter).
		WithStartFunctions())
	if err != nil {
		return nil, err
	}
	defer func() {
		if !ok {
			e.Close()
		}
	}()

	proc := &Proc{
		Sys:      sys,
		Module:   mod,
		Endpoint: e}
	c.Host.SetStreamHandler(e.Protocol(), func(s network.Stream) {
		defer s.Close()

		if err := proc.Poll(ctx, s, nil); err != nil {
			slog.ErrorContext(ctx, "failed to poll process", "reason", err)
			return
		}
	})
	runtime.SetFinalizer(proc, func(p *Proc) {
		c.Host.RemoveStreamHandler(e.Protocol())
	})
	ok = true
	return proc, nil
}

type Proc struct {
	Sys      api.Closer
	Endpoint *Endpoint
	Module   api.Module
}

// ID returns the process identifier (endpoint name) without the protocol prefix.
func (p *Proc) ID() string {
	return p.Endpoint.Name
}

func (p *Proc) Close(ctx context.Context) error {
	return multierr.Combine(
		p.Endpoint.Close(),
		p.Module.Close(ctx),
		p.Sys.Close(ctx))
}

func (p Proc) Poll(ctx context.Context, s network.Stream, stack []uint64) error {
	if deadline, ok := ctx.Deadline(); ok {
		if err := s.SetReadDeadline(deadline); err != nil {
			return fmt.Errorf("set read deadline: %w", err)
		}
	}

	p.Endpoint.ReadWriteCloser = s
	defer func() {
		p.Endpoint.ReadWriteCloser = nil
	}()

	if poll := p.Module.ExportedFunction("poll"); poll == nil {
		return fmt.Errorf("%s::poll: not found", p)
	} else {
		return poll.CallWithStack(ctx, stack)
	}
}
