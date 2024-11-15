package proc

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"runtime"

	"capnproto.org/go/capnp/v3"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type Config struct {
	Args, Env      []string
	Stdin          io.Reader
	Stdout, Stderr io.Writer
}

func (cfg Config) Instantiate(
	ctx context.Context,
	r wazero.Runtime,
	cm wazero.CompiledModule,
) (*P, error) {
	var p P

	// /ww/<semver>/proc/<pid>
	pid := NewPID().String()

	var err error
	p.Mod, err = r.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
		WithName(pid).
		WithArgs(cfg.Args...).
		WithStdin(&p.mailbox).
		WithStdout(cfg.Stdout).
		WithStderr(cfg.Stderr).
		WithEnv("WW_PID", pid).
		WithRandSource(rand.Reader).
		WithOsyield(runtime.Gosched).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithStartFunctions())
	return &p, err
}

type P struct {
	Parent protocol.ID
	Mod    api.Module

	stack   []uint64
	mailbox bytes.Reader
}

func (p *P) String() string {
	return p.Mod.Name()
}

func (p *P) Close(ctx context.Context) error {
	return p.Mod.Close(ctx)
}

func (p *P) Deliver(ctx context.Context, call MethodCall) error {
	// defensive; underlying slice owned by 'call', becomes invalid
	// after caller releases the underlying capnp.Message.
	defer p.mailbox.Reset(nil)

	name, err := call.Name()
	if err != nil {
		return err
	}
	fn := p.Mod.ExportedFunction(name)
	if fn == nil {
		return errors.New("missing export: " + name)
	}

	stack, err := call.Stack()
	if err != nil {
		return err
	}

	data, err := call.CallData()
	if err != nil {
		return err
	}
	p.mailbox.Reset(data) // reset stdin

	err = fn.CallWithStack(ctx, p.ToWasmStack(stack))
	if errors.Is(err, context.Canceled) {
		err = context.Canceled
	} else if errors.Is(err, context.DeadlineExceeded) {
		err = context.DeadlineExceeded
	}

	return err
}

func (p *P) ToWasmStack(stack capnp.UInt64List) []uint64 {
	if stack.Len() <= 0 {
		return nil
	} else if cap(p.stack) < stack.Len() {
		p.stack = make([]uint64, stack.Len())
	} else {
		p.stack = p.stack[:stack.Len()]
	}

	for i := 0; i < stack.Len(); i++ {
		p.stack[i] = stack.At(i)
	}

	return p.stack
}
