package proc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"path"
	"strings"

	"capnproto.org/go/capnp/v3"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type Config struct {
	Proto          protocol.ID
	Args, Env      []string
	Stdout, Stderr io.Writer
	Runtime        wazero.Runtime
	Module         wazero.CompiledModule
}

func (cfg Config) Bind(ctx context.Context, p *P) (err error) {
	pid := NewPID()
	proto := path.Join(string(cfg.Proto), pid.String()) // /ww/0.1.0/<pid>

	mc := wazero.NewModuleConfig().
		WithName(pid.String()).
		WithArgs(cfg.Args...).
		WithStdin(&p.mailbox).
		WithStdout(cfg.Stdout).
		WithStderr(cfg.Stderr).
		WithEnv("WW_PROTO", proto)

	p.Mod, err = cfg.Runtime.InstantiateModule(ctx, cfg.Module,
		cfg.WithEnv(mc))
	return
}

func (cfg Config) WithEnv(mc wazero.ModuleConfig) wazero.ModuleConfig {
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
	Mod api.Module

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

// rd      io.LimitedReader
// func (p *P) ReadFrom(r io.Reader) (int64, error) {
// 	p.buf.Reset()
// 	p.rd.R = r
// 	p.rd.N = 1<<32 - 1 // max uint32
// 	return io.Copy(&p.buf, r)
// }
