package proc

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"log/slog"
	"runtime"
	"strings"

	"capnproto.org/go/capnp/v3"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type Command struct {
	Args, Env      []string
	Stdout, Stderr io.Writer
}

func (cmd Command) Instantiate(ctx context.Context, r wazero.Runtime, cm wazero.CompiledModule) (*P, error) {
	var p P
	var err error

	// /ww/<semver>/proc/<pid>
	pid := NewPID().String()
	config := cmd.WithEnv(wazero.NewModuleConfig().
		WithName(pid).
		WithArgs(cmd.Args...).
		WithStdin(&p.mailbox).
		WithStdout(cmd.Stdout).
		WithStderr(cmd.Stderr).
		WithEnv("WW_PID", pid).
		WithRandSource(rand.Reader).
		WithOsyield(runtime.Gosched).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithStartFunctions())
	p.Mod, err = r.InstantiateModule(ctx, cm, config)
	return &p, err
}

func (cfg Command) WithEnv(mc wazero.ModuleConfig) wazero.ModuleConfig {
	for _, s := range cfg.Env {
		ss := strings.SplitN(s, "=", 2)
		if len(ss) != 2 {
			slog.Warn("ignored unparsable environment variable",
				"var", s)
			continue
		}

		mc = mc.WithEnv(ss[0], ss[1])
	}

	return mc
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
