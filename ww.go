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
	guest "github.com/wetware/go/std/system"
	"github.com/wetware/go/system"

	protoutils "github.com/wetware/go/util/proto"
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
	}

	if b, err := io.ReadAll(env.IO.Stdin); err != nil {
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
		case guest.StatusAsync:
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

	return proc.Command{
		Args:   env.IO.Args,
		Env:    env.IO.Env,
		Stdout: env.IO.Stdout,
		Stderr: env.IO.Stderr,
	}.Instantiate(ctx, r, cm)
}

func (env Env) LoadAndCompile(ctx context.Context, r wazero.Runtime, name string) (wazero.CompiledModule, error) {
	p, err := path.NewPath(name)
	if err != nil {
		return nil, err
	}

	n, err := env.FS.Unix.Get(ctx, p)
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
				child := filepath.Join(p.String(), it.Name())
				return env.LoadAndCompile(ctx, r, child)
			}
		}
	}

	return nil, errors.New("binary not found")
}
