//go:generate mockgen -source=proc.go -destination=mock_test.go -package=proc_test

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
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"golang.org/x/sync/semaphore"
)

var ErrMethodNotFound = errors.New("method not found")

type Method interface {
	CallWithStack(context.Context, []uint64) error
}

type Command struct {
	PID       ID
	Args, Env []string
	Stderr    io.Writer
	FS        fs.FS
}

// Module represents the subset of wazero.Module methods we use
type Module interface {
	Name() string
	Close(context.Context) error
	ExportedFunction(string) api.Function
}

func (cmd Command) Instantiate(ctx context.Context, r wazero.Runtime, cm wazero.CompiledModule) (*P, error) {
	var p P
	var err error
	p.Mod, err = r.InstantiateModule(ctx, cm, cmd.WithEnv(wazero.NewModuleConfig().
		WithName(cmd.PID.String()).
		WithArgs(cmd.Args...).
		WithStdin(&p.Mailbox).
		WithStdout(&p.SendQueue).
		WithStderr(cmd.Stderr).
		WithEnv("WW_PID", cmd.PID.String()).
		WithFS(cmd.FS).
		WithRandSource(rand.Reader).
		WithOsyield(runtime.Gosched).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithStartFunctions()))
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
	Mailbox   struct{ io.Reader }
	SendQueue struct{ io.Writer }
	Mod       Module

	once sync.Once
	sem  *semaphore.Weighted
}

func (p *P) String() string {
	return p.Mod.Name()
}

func (p *P) Close(ctx context.Context) error {
	return p.Mod.Close(ctx)
}

func (p *P) Reserve(ctx context.Context, conn io.ReadWriteCloser) error {
	p.once.Do(func() {
		p.sem = semaphore.NewWeighted(1)
	})

	err := p.sem.Acquire(ctx, 1)
	if err == nil {
		p.SendQueue.Writer = conn
		p.Mailbox.Reader = conn
	}

	return err
}

func (p *P) Release() {
	defer p.sem.Release(1)
	p.Mailbox.Reader = nil
	p.SendQueue.Writer = nil
}

func (p *P) Method(name string) Method {
	return p.Mod.ExportedFunction(name)
}
