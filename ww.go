package ww

import (
	"context"
	"io"

	"capnproto.org/go/capnp/v3"
	"github.com/blang/semver/v4"
	"github.com/tetratelabs/wazero"
	"github.com/wetware/go/proc"
	guest "github.com/wetware/go/std/system"
	"github.com/wetware/go/system"

	protoutils "github.com/wetware/go/util/proto"
)

// These special exit codes are reserved by Wetware.  It is assumed that a well-behaved WASM program
// will not return these exit codes.  In general, we assume that the most-significant 16 bits are reserved
// for Wetware.
const (
	// ExitCodePivot
	ExitCodePivot uint32 = 0x00ff0000
)

const Version = "0.1.0"

var Proto = protoutils.VersionedID{
	ID:      "ww",
	Version: semver.MustParse(Version),
}

type Env struct {
	IO  system.IO
	Net system.Net
	FS  system.IPFS
}

func (env Env) Bind(ctx context.Context, r wazero.Runtime) error {
	p, err := env.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer p.Close(ctx)

	release := env.Net.Bind(ctx, p)
	defer release()

	// Call main() function (alias _start method)
	m, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return err
	}
	defer m.Release()

	call, err := proc.NewRootMethodCall(seg)
	if err != nil {
		return err
	} else if err := call.SetName("_start"); err != nil {
		return err
	} else if b, err := io.ReadAll(env.IO.Stdin); err != nil {
		return err
	} else if err := call.SetCallData(b); err != nil {
		return err
	}

	err = p.Deliver(ctx, call)
	switch e := err.(type) {
	case interface{ ExitCode() uint32 }:
		switch e.ExitCode() {
		case 0:
			return nil
		case guest.StatusAwaiting:
			return env.Net.ServeProc(ctx, p)
		}
	}

	return err
}

func (env Env) Instantiate(ctx context.Context, r wazero.Runtime) (*proc.P, error) {
	cm, err := env.LoadAndCompile(ctx, r, env.IO.Args[0]) // FIXME:  panic if len(args)=0
	if err != nil {
		return nil, err
	}
	defer cm.Close(ctx)

	return proc.Config{
		Args:   env.IO.Args,
		Env:    env.IO.Env,
		Stdin:  env.IO.Stdin,
		Stdout: env.IO.Stdout,
		Stderr: env.IO.Stderr,
	}.Instantiate(ctx, r, cm)
}

func (env Env) LoadAndCompile(ctx context.Context, r wazero.Runtime, name string) (wazero.CompiledModule, error) {
	b, err := env.ReadAll(ctx, name)
	if err != nil {
		return nil, err
	}

	return r.CompileModule(ctx, b)
}

func (env Env) ReadAll(ctx context.Context, name string) ([]byte, error) {
	f, err := env.FS.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

// func (env Env) WithEnv(mc wazero.ModuleConfig) wazero.ModuleConfig {
// 	for _, s := range env.IO.Env {
// 		ss := strings.SplitN(s, "=", 2)
// 		if len(ss) != 2 {
// 			slog.Warn("ignored unparsable environment variable",
// 				"var", s)
// 			continue
// 		}

// 		mc = mc.WithEnv(ss[0], ss[1])
// 	}

// 	return mc
// }
