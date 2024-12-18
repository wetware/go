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
	PID            PID
	Args, Env      []string
	Stdout, Stderr io.Writer
	FS             fs.FS
}

func (cmd Command) Instantiate(ctx context.Context, r wazero.Runtime, cm wazero.CompiledModule) (*P, error) {
	var p P
	var err error
	p.Mod, err = r.InstantiateModule(ctx, cm, cmd.WithEnv(wazero.NewModuleConfig().
		WithName(cmd.PID.String()).
		WithArgs(cmd.Args...).
		WithStdin(&p.Mailbox).
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
	Mailbox struct{ io.Reader }
	Mod     api.Module

	once sync.Once
	sem  *semaphore.Weighted
}

func (p *P) String() string {
	return p.Mod.Name()
}

func (p *P) Close(ctx context.Context) error {
	return p.Mod.Close(ctx)
}

func (p *P) Reserve(ctx context.Context, body io.Reader) error {
	p.once.Do(func() {
		p.sem = semaphore.NewWeighted(1)
	})

	return p.sem.Acquire(ctx, 1)
}

func (p *P) Release() {
	defer p.sem.Release(1)
	p.Mailbox.Reader = nil
}

func (p *P) Method(name string) Method {
	return p.Mod.ExportedFunction(name)
}

// func (p *P) Deliver(ctx context.Context, body io.Reader) ([]uint64, error) {
// 	// Acquire a lock on the process
// 	////
// 	err := p.sem.Acquire(ctx, 1)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer p.sem.Release(-1)

// 	// Assign stream to mailbox and call method with stack.
// 	////
// 	p.Mailbox.Reader = body                   // we found the method; last chance to hook in
// 	defer func() { p.Mailbox.Reader = nil }() // defensive; catch use-after-free

// 	err = method.CallWithStack(ctx, stack)
// 	if errors.Is(err, context.Canceled) {
// 		err = context.Canceled
// 	} else if errors.Is(err, context.DeadlineExceeded) {
// 		err = context.DeadlineExceeded
// 	}

// 	return stack, err
// }

// func (p *P) ReadCallData(r io.Reader) (method api.Function, stack []uint64, err error) {
// 	rd := bufio.NewReader(r)
// 	if method, err = p.ReadAndLoadMethod(rd); err != nil {
// 		return
// 	} else if method == nil {
// 		err = ErrMethodNotFound
// 		return
// 	}

// 	stack, err = ReadStack(rd)
// 	return
// }

// func (p *P) ReadAndLoadMethod(rd io.ByteReader) (api.Function, error) {
// 	// Length prefix
// 	n, err := binary.ReadUvarint(rd)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var b byte
// 	var s strings.Builder
// 	for i := uint64(0); i < n; i++ {
// 		if b, err = rd.ReadByte(); err != nil {
// 			break
// 		} else if err = s.WriteByte(b); err != nil {
// 			break
// 		}
// 	}

// 	method := p.Mod.ExportedFunction(s.String())
// 	return method, err
// }

// func ReadStack(rd io.ByteReader) ([]uint64, error) {
// 	// Length prefix
// 	n, err := binary.ReadUvarint(rd)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Read n words from the stack
// 	var word uint64
// 	var stack []uint64
// 	for i := uint64(0); i < n; i++ {
// 		if word, err = binary.ReadUvarint(rd); err != nil {
// 			break
// 		}

// 		stack = append(stack, word)
// 	}

// 	return stack, err
// }
