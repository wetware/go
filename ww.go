package ww

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tetratelabs/wazero"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
)

type ExitError interface {
	error
	ExitCode() uint32
}

type Env struct {
	IPFS iface.CoreAPI
	Host host.Host
	Cmd  system.Cmd
	Net  system.Net
	FS   system.Anchor
}

func (env Env) Bind(ctx context.Context, r wazero.Runtime) error {
	cm, err := env.LoadAndCompile(ctx, r, env.Cmd.Args[0]) // FIXME:  panic if len(args)=0
	if err != nil {
		return err
	}
	defer cm.Close(ctx)

	p, err := env.Instantiate(ctx, r, cm)
	if err != nil {
		return err
	}
	defer p.Close(ctx)

	release, err := env.Net.Bind(ctx, p /* TODO:  pass in 'ns' here */)
	if err != nil {
		return err
	}
	defer release()

	call := &proc.Call{Method: "_start"}
	err = p.Deliver(ctx, call, env.Cmd.Stdin)
	switch e := err.(type) {
	case ExitError:
		if e.ExitCode() == 0 {
			err = nil
		}
	}

	if err != nil {
		return err
	}

	return env.Net.ServeProc(ctx, p)
}

func (env Env) Instantiate(ctx context.Context, r wazero.Runtime, cm wazero.CompiledModule) (*proc.P, error) {
	return proc.Command{
		PID:    proc.NewPID(),
		Args:   env.Cmd.Args,
		Env:    env.Cmd.Env,
		Stdout: env.Cmd.Stdout,
		Stderr: env.Cmd.Stderr,
		FS:     env.FS,
	}.Instantiate(ctx, r, cm)
}

func (env Env) LoadAndCompile(ctx context.Context, r wazero.Runtime, name string) (wazero.CompiledModule, error) {
	p, err := path.NewPath(name)
	if err != nil {
		return nil, err
	}

	n, err := env.OpenUnix(ctx, p)
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

func (env Env) OpenUnix(ctx context.Context, p path.Path) (files.Node, error) {
	root := env.IPFS.Unixfs()
	return root.Get(ctx, p)
}

func (env Env) ProtoFor(pid fmt.Stringer) protocol.ID {
	proto := filepath.Join(system.Proto.String(), "proc", pid.String())
	return protocol.ID(proto)
}
