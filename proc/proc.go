package proc

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"runtime"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"golang.org/x/sync/semaphore"
)

type Command struct {
	PID            PID
	Args, Env      []string
	Stdout, Stderr io.Writer
	FS             fs.FS
}

func (cmd Command) Instantiate(ctx context.Context, r wazero.Runtime, cm wazero.CompiledModule) (*P, error) {
	mod, err := r.InstantiateModule(ctx, cm, cmd.WithEnv(wazero.NewModuleConfig().
		WithName(cmd.PID.String()).
		WithArgs(cmd.Args...).
		WithStdout(cmd.Stdout).
		WithStderr(cmd.Stderr).
		WithEnv("WW_PID", cmd.PID.String()).
		WithFS(cmd.FS).
		WithRandSource(rand.Reader).
		WithOsyield(runtime.Gosched).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithStartFunctions()))
	return New(mod), err
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
	mod     api.Module
	mailbox struct{ io.Reader }
	sem     *semaphore.Weighted
}

func New(mod api.Module) *P {
	return &P{
		mod: mod,
		sem: semaphore.NewWeighted(1),
	}
}

func (p *P) String() string {
	return p.mod.Name()
}

func (p *P) Close(ctx context.Context) error {
	return p.mod.Close(ctx)
}

func (p *P) Deliver(ctx context.Context, call *Call, r io.Reader) error {
	p.mailbox.Reader = r                      // we found the method; last chance to hook in
	defer func() { p.mailbox.Reader = nil }() // defensive; catch use-after-free

	return call.Eval(ctx, p.mod)
}

type Call struct {
	Method string
	Stack  []uint64
}

func ParseCallData(s string) (*Call, error) {

}

func (call Call) Eval(ctx context.Context, mod api.Module) error {
	fn := mod.ExportedFunction(call.Method)
	if fn == nil {
		return errors.New("missing export: " + call.Method)
	}

	err := fn.CallWithStack(ctx, call.Stack)
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	} else if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}

	return err
}
