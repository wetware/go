package vat

import (
	"context"
	"log/slog"
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type ReleaseFunc func()

type StreamHandler struct {
	Host   host.Host
	Proto  protocol.ID
	accept <-chan network.Stream
}

func (h *StreamHandler) Bind(ctx context.Context) ReleaseFunc {
	ch := make(chan network.Stream)
	*h = StreamHandler{
		Host:   h.Host,
		Proto:  h.Proto,
		accept: ch,
	}

	h.Host.SetStreamHandler(h.Proto, h.NewStreamHandler(ctx, ch))
	return h.NewRelease(func() {
		defer close(ch)
		defer h.Host.RemoveStreamHandler(h.Proto)
	})
}

func (h *StreamHandler) NewStreamHandler(ctx context.Context, ch chan<- network.Stream) network.StreamHandler {
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

func (h *StreamHandler) NewRelease(f func()) ReleaseFunc {
	var once sync.Once
	return func() {
		once.Do(f) // f is called at most once.
	}
}

func (h *StreamHandler) Accept(ctx context.Context) (s network.Stream, err error) {
	select {
	case s = <-h.accept:
	case <-ctx.Done():
		err = ctx.Err()
	}

	return
}
