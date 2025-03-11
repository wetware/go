package glia

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/wetware/go/system"
)

var ErrU16Overflow = errors.New("varint overflows u16")

type P2P struct {
	Env    Env
	Router system.Router
}

func (p2p *P2P) Log() *slog.Logger {
	return p2p.Env.Log().With(
		"service", p2p.String())
}

func (p2p *P2P) String() string {
	return "p2p"
}

func (p2p *P2P) Serve(ctx context.Context) error {
	proto := system.Proto.Unwrap()
	p2p.Env.LocalHost().SetStreamHandlerMatch(proto,
		func(id protocol.ID) bool {
			root := system.Proto.Path()
			return strings.HasPrefix(string(id), root)
		},
		func(s network.Stream) {
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
	defer p2p.Env.LocalHost().RemoveStreamHandler(proto)
	p2p.Log().DebugContext(ctx, "service started")

	// If we made it this far, we're bootstrapped.  Wait for shutdown
	// signal.
	<-ctx.Done()
	return ctx.Err()
}

func (p2p *P2P) ServeStream(ctx context.Context, s Stream) error {
	defer s.Close()
	// Glia RPC is a synchronous RPC protocol models one round-trip
	// (request-response) between a server and a client.  The round-
	// trip models a synchronous method call on an object.
	////

	// Local call?
	////
	if p2p.Env.LocalHost().ID().String() == s.Destination() {
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

	// Forward the call
	////
	dst, err := peer.Decode(s.Destination())
	if err != nil {
		return err
	}
	remote, err := p2p.Env.LocalHost().NewStream(ctx, dst, s.Protocol())
	if err != nil {
		return err
	}
	defer s.Close()

	// Forward the request stream
	if _, err := io.Copy(remote, s); err != nil {
		return err
	}
	if err := remote.CloseWrite(); err != nil {
		return err
	}

	// Read back the response stream
	if _, err := io.Copy(s, remote); err != nil {
		return err
	}
	if err := remote.CloseRead(); err != nil {
		return err
	}

	return nil
}

type P2PStream struct {
	network.Stream
}

var _ Stream = (*P2PStream)(nil)

func (s P2PStream) Close() error {
	return s.Stream.Close()
}

func (s P2PStream) Destination() string {
	proto := s.Protocol()
	p := path.Dir(string(proto))
	p = path.Dir(p)
	return path.Base(p)
}

func (s P2PStream) ProcID() string {
	proto := s.Protocol()
	dir := path.Dir(string(proto))
	return path.Base(dir)
}

func (s P2PStream) MethodName() string {
	proto := s.Protocol()
	return path.Base(string(proto))

}
