package ww

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"capnproto.org/go/capnp/v3"
	"github.com/blang/semver/v4"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/tetratelabs/wazero"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"

	protoutils "github.com/wetware/go/util/proto"
)

const Version = "0.1.0"

var Proto = protoutils.VersionedID{
	ID:      "ww",
	Version: semver.MustParse(Version),
}

type Env struct {
	Cmd system.Cmd
	Net system.Net
	FS  system.FS
}

func (env Env) Bind(ctx context.Context, r wazero.Runtime) error {
	cm, err := env.LoadAndCompile(ctx, r, env.Cmd.Path) // FIXME:  panic if len(args)=0
	if err != nil {
		return err
	}
	defer cm.Close(ctx)

	p, err := env.Instantiate(ctx, r, cm)
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
	}

	if b, err := io.ReadAll(env.Cmd.Stdin); err != nil {
		return err
	} else if err := call.SetCallData(b); err != nil {
		return err
	}

	err = p.Deliver(ctx, call)
	if e, ok := err.(system.ExitError); ok && e.ExitCode() != 0 {
		return err
	} else if err != nil {
		return err
	}

	return env.Net.ServeProc(ctx, p)
}

func (env Env) Instantiate(ctx context.Context, r wazero.Runtime, cm wazero.CompiledModule) (*proc.P, error) {
	return proc.Command{
		Path:   env.Cmd.Path,
		Args:   env.Cmd.Args,
		Env:    env.Cmd.Env,
		Stdout: env.Cmd.Stdout,
		Stderr: env.Cmd.Stderr,
	}.Instantiate(ctx, r, cm)
}

func (env Env) LoadAndCompile(ctx context.Context, r wazero.Runtime, name string) (wazero.CompiledModule, error) {
	p, err := path.NewPath(name)
	if err != nil {
		return nil, err
	}

	n, err := env.FS.OpenUnix(ctx, p)
	if err != nil {
		return nil, err
	}
	defer n.Close()

	switch node := n.(type) {
	case files.File:
		b, err := io.ReadAll(node)
		if err != nil {
			return nil, err
		}
		return r.CompileModule(ctx, b)

	case files.Directory:
		it := node.Entries()
		for it.Next() {
			if it.Name() == "main.wasm" {
				child := filepath.Join(name, it.Name())
				return env.LoadAndCompile(ctx, r, child)
			}
		}
	}

	return nil, errors.New("binary not found")
}
