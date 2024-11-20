package system

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"strings"

	"capnproto.org/go/capnp/v3"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/wetware/go/proc"
)

type Handler interface {
	ServeProc(context.Context, *proc.P) error
}

type Net struct {
	Proto protocol.ID
	Host  host.Host
	Handler
}

func (n Net) Match(id protocol.ID) bool {
	proto := strings.TrimPrefix(string(id), string(n.Proto))
	return strings.HasPrefix(proto, "/proc/")
}

func (n Net) Bind(ctx context.Context, p *proc.P) network.StreamHandler {
	log := slog.Default().With(
		"peer", n.Host.ID(),
		"pid", p.String())

	return func(s network.Stream) {
		defer s.Close()

		// TODO:  handle context deadline

		if call, err := ReadCall(s); err != nil {
			log.ErrorContext(ctx, "failed to read method call",
				"reason", err)
		} else if err := p.Deliver(ctx, call); err != nil {
			log.ErrorContext(ctx, "failed to deliver method call",
				"reason", err)
		}
	}
}

func (n Net) ServeProc(ctx context.Context, p *proc.P) error {
	if n.Handler == nil {
		return nil
	}

	return n.Handler.ServeProc(ctx, p)
}

func ReadCall(r io.Reader) (proc.MethodCall, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return proc.MethodCall{}, err
	}

	m, err := capnp.Unmarshal(b)
	if err != nil {
		return proc.MethodCall{}, err
	}

	return proc.ReadRootMethodCall(m)
}

// HostNode allows
type HostNode struct {
	Ctx  context.Context
	Host host.Host
}

func (h HostNode) Open(name string) (fs.File, error) {
	path, err := NewPath(name)
	if err != nil {
		return nil, err
	}

	return h.Walk(h.Ctx, path)
}

func (h HostNode) Walk(ctx context.Context, p Path) (fs.File, error) {
	/*
		Example path:  /ww/0.1.0/proc/<pid>
	*/

	id, err := p.Peer()
	if err != nil {
		return nil, err
	}

	proto, err := p.Proto()
	if err != nil {
		return nil, err
	}

	s, err := h.Host.NewStream(ctx, id, proto)
	return StreamNode{Stream: s}, err
}

type HandlerFunc func(context.Context, *proc.P) error

func (serve HandlerFunc) ServeProc(ctx context.Context, p *proc.P) error {
	return serve(ctx, p)
}
