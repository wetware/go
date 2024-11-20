package ww

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"capnproto.org/go/capnp/v3"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tetratelabs/wazero"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
)

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

	// Bind libp2p streams that allow remote peers to send
	// messages to p.
	env.Host.SetStreamHandlerMatch(env.ProtoFor(p),
		env.Net.Match,
		env.Net.Bind(ctx, p))
	defer env.Host.RemoveStreamHandler(env.ProtoFor(p))
	slog.DebugContext(ctx, "attached process stream handlers")

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
	switch e := err.(type) {
	case system.ExitError:
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
