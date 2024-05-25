package vat

import (
	"context"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type ListenConfig struct {
	Host host.Host
}

func (c ListenConfig) Listen(ctx context.Context, id protocol.ID) Listener {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan network.Stream)
	c.Host.SetStreamHandler(id, NewStreamHandler(ctx, ch))

	return Listener{
		C: ch,
		Release: func() {
			defer close(ch)
			defer cancel()
			c.Host.RemoveStreamHandler(id)
		},
	}
}

func NewStreamHandler(ctx context.Context, ch chan<- network.Stream) network.StreamHandler {
	return func(s network.Stream) {
		select {
		case ch <- s:
			slog.DebugContext(ctx, "handled stream",
				"peer", s.Conn().RemotePeer(),
				"stream", s.ID(),
				"protocol", s.Protocol())

		case <-ctx.Done():
			if err := s.Close(); err != nil {
				slog.ErrorContext(ctx, "stream close failed",
					"reason", err,
					"peer", s.Conn().RemotePeer(),
					"stream", s.ID(),
					"protocol", s.Protocol())
			}
		}
	}
}

type Listener struct {
	C       <-chan network.Stream
	Release func()
}

func (h Listener) Accept(ctx context.Context) (s network.Stream, err error) {
	select {
	case s = <-h.C:
	case <-ctx.Done():
		err = ctx.Err()
	}

	return
}
