package system

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/mr-tron/base58"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
	"go.uber.org/multierr"
	"golang.org/x/sync/semaphore"
)

type ProcConfig struct {
	IPFS      IPFS
	Host      host.Host
	Runtime   wazero.Runtime
	Src       io.ReadCloser
	Env, Args []string
	ErrWriter io.Writer
	Async     bool // If true, use WithStartFunctions() and set up stream handler
}

func (c ProcConfig) New(ctx context.Context) (*Proc, error) {
	var ok = false
	var cs CloserSlice
	defer func() {
		if !ok {
			cs.Close(ctx)
		}
	}()

	if c.Src == nil {
		return nil, errors.New("no source provided")
	}

	bytecode, err := io.ReadAll(c.Src)
	if err != nil {
		return nil, err
	}
	defer c.Src.Close()

	cm, err := c.Runtime.CompileModule(ctx, bytecode)
	if err != nil {
		return nil, err
	}
	cs = append(cs, cm)

	wasi, err := wasi_snapshot_preview1.Instantiate(ctx, c.Runtime)
	if err != nil {
		return nil, err
	}
	cs = append(cs, wasi)

	e := c.NewEndpoint()
	cs = append(cs, e)

	// In sync mode, set up stdin/stdout for the endpoint
	if !c.Async {
		// bidirectional pipe that wraps stdin/stdout.
		e.ReadWriteCloser = struct {
			io.Reader
			io.WriteCloser
		}{
			Reader:      os.Stdin,
			WriteCloser: os.Stdout,
		}
	}

	// Configure module instantiation based on async mode
	config := c.NewModuleConfig(ctx, e)

	mod, err := c.Runtime.InstantiateModule(ctx, cm, config)
	if err != nil {
		// Check if the error is sys.ExitError with exit code 0 which indicates success
		var exitErr *sys.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 0 {
			// Exit code 0 means success, so we can continue
			err = nil
		} else {
			return nil, err
		}
	}

	// Check if module is closed after instantiation
	// In sync mode, this is expected behavior as main() completes and exits
	if mod.IsClosed() && !c.Async {
		// In sync mode, module closure after successful execution is normal
		// We'll create a minimal proc that can still be used for ID purposes
	} else if mod.IsClosed() {
		return nil, fmt.Errorf("module closed immediately after instantiation")
	}
	cs = append(cs, mod)

	// Mark proc as initialized and optionally bind stream handler.
	////
	ok = true
	proc := &Proc{
		Config:   c,
		Module:   mod,
		Endpoint: e,
		Closer:   cs,
		sem:      semaphore.NewWeighted(1)}
	return proc, nil
}

type ReadWriteStringer interface {
	String() string
	io.ReadWriter
}

func (c ProcConfig) NewModuleConfig(ctx context.Context, sock ReadWriteStringer) wazero.ModuleConfig {
	config := wazero.NewModuleConfig().
		WithName(sock.String()).
		WithArgs(c.Args...).
		WithStdin(sock).
		WithStdout(sock).
		WithStderr(c.ErrWriter).
		WithFSConfig(c.NewFSConfig(ctx))

	// async mode?
	if c.Async {
		// prevent _start from running automatically
		config = config.WithStartFunctions()
	}

	// Add environment variables
	for _, env := range c.Env {
		if k, v, ok := strings.Cut(env, "="); ok {
			config = config.WithEnv(k, v)
		}
	}

	return config
}

func (c ProcConfig) NewFSConfig(ctx context.Context) wazero.FSConfig {
	ipfs := IPFS{
		Ctx:  ctx,
		Unix: c.IPFS.Unix,
		Root: c.IPFS.Root}

	return wazero.NewFSConfig().
		WithFSMount(ipfs, "/ipfs").
		WithFSMount(ipfs, "/ipns").
		WithFSMount(ipfs, "/ipld")
}

func (p ProcConfig) NewEndpoint() *Endpoint {
	var buf [8]byte
	if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
		panic(err)
	}

	return &Endpoint{Name: base58.FastBase58Encoding(buf[:])}
}

type Proc struct {
	Config   ProcConfig
	Endpoint *Endpoint
	Module   api.Module
	api.Closer

	sem *semaphore.Weighted
}

// ID returns the process identifier (endpoint name) without the protocol prefix.
func (p Proc) ID() string {
	return p.Endpoint.Name
}

// ProcessMessage processes one complete message synchronously.
// In sync mode: lets _start run automatically and process one message
// In async mode: calls the specified export function
func (p Proc) ProcessMessage(ctx context.Context, s network.Stream, method string) error {
	if deadline, ok := ctx.Deadline(); ok {
		if err := s.SetReadDeadline(deadline); err != nil {
			return fmt.Errorf("set read deadline: %w", err)
		}
	}

	// Check if module is closed before we start
	if p.Module.IsClosed() {
		return fmt.Errorf("%s::ProcessMessage: module closed", p.ID())
	}

	// Set the stream as the endpoint's ReadWriteCloser for this message
	// The Endpoint's Read/Write methods will delegate to the stream
	p.Endpoint.ReadWriteCloser = s
	defer func() {
		// Reset to nil after processing this message
		p.Endpoint.ReadWriteCloser = nil
	}()

	// In async mode, call the specified export function
	if p.Config.Async {
		// Normalize method: if empty string, use "poll"
		if method == "" {
			method = "poll"
		}

		exp := p.Module.ExportedFunction(method)
		if exp == nil {
			_ = s.Reset()
			return fmt.Errorf("unknown method: %s", method)
		}

		if err := p.sem.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("acquire semaphore: %w", err)
		}
		defer p.sem.Release(1)

		if err := exp.CallWithStack(ctx, nil); err != nil {
			var exitErr *sys.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() != 0 {
				return fmt.Errorf("%s::%s: %w", p.ID(), method, err)
			}
			// If it's ExitError with code 0, treat as success
		}
	}
	// In sync mode, _start already ran during module instantiation

	return nil
}

type CloserSlice []api.Closer

func (cs CloserSlice) Close(ctx context.Context) error {
	var errs []error
	for _, c := range cs {
		errs = append(errs, c.Close(ctx))
	}
	return multierr.Combine(errs...)
}
