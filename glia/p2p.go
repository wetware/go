package glia

import (
	"context"
	"errors"
	"log/slog"
	"path"

	"github.com/libp2p/go-libp2p/core/network"

	"github.com/wetware/go/system"
)

var ErrU16Overflow = errors.New("varint overflows u16")

type P2P struct {
	Env    *system.Env
	Router Router
}

func (p2p P2P) Log() *slog.Logger {
	return p2p.Env.Log().With(
		"service", p2p.String())
}

func (p2p P2P) String() string {
	return "p2p"
}

func (p2p P2P) Serve(ctx context.Context) error {
	proto := system.Proto.Unwrap()
	p2p.Env.Host.SetStreamHandler(proto, func(s network.Stream) {
		defer s.Close()

		if dl, ok := ctx.Deadline(); ok {
			if err := s.SetDeadline(dl); err != nil {
				p2p.Log().WarnContext(ctx, "failed to set deadline",
					"reason", err)
				// non-fatal; continue along...
			}
		}

		if err := p2p.ServeStream(ctx, P2PStream{Stream: s}); err != nil {
			p2p.Log().ErrorContext(ctx, "failed to serve stream",
				"reason", err,
				"stream", s.ID())
		}
	})
	defer p2p.Env.Host.RemoveStreamHandler(proto)
	p2p.Log().DebugContext(ctx, "service started")

	<-ctx.Done()
	return nil
}

func (p2p P2P) ServeStream(ctx context.Context, s Stream) error {
	defer s.Close()
	// Glia RPC is a synchronous RPC protocol models one round-trip
	// (request-response) between a server and a client.  The round-
	// trip models a synchronous method call on an object.
	////

	p, err := p2p.Router.GetProc(s.ProcID())
	if err != nil {
		return err
	}

	if err := p.Reserve(ctx, s); err != nil {
		return err
	}
	defer p.Release()

	return p.Method(s.MethodName()).CallWithStack(ctx, nil) // TODO:  stack
}

type P2PStream struct {
	network.Stream
}

var _ Stream = (*P2PStream)(nil)

// func (s P2PStream) Host() string {
// 	return s.Conn().RemotePeer().String()
// }

func (s P2PStream) ProcID() string {
	proto := s.Protocol()
	dir := path.Dir(string(proto))
	if dir == "." {
		return ""
	}
	return path.Base(dir)
}

func (s P2PStream) MethodName() string {
	proto := s.Protocol()
	return path.Base(string(proto))

}
