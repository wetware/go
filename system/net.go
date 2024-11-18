package system

import (
	"context"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/wetware/go/proc"
	protoutils "github.com/wetware/go/util/proto"
)

type Net struct {
	Proto protoutils.VersionedID
	Host  host.Host
	Handler
}

func (n Net) Bind(ctx context.Context, p *proc.P) ReleaseFunc {
	handler := proc.StreamHandler{
		VersionedID: n.Proto,
		Proc:        p}
	proto := handler.Proto()
	pid := handler.String()
	peer := n.Host.ID()

	n.Host.SetStreamHandlerMatch(
		proto,
		handler.Match,
		handler.Bind(ctx))
	slog.DebugContext(ctx, "attached process stream handlers",
		"peer", peer,
		"proto", proto,
		"proc", pid)
	return func() {
		n.Host.RemoveStreamHandler(proto)
		slog.DebugContext(ctx, "detached process stream handlers",
			"peer", peer,
			"proto", proto,
			"proc", pid)
	}
}

func (n Net) ServeProc(ctx context.Context, p *proc.P) (err error) {
	if n.Handler != nil {
		err = n.Handler.ServeProc(ctx, p)
	}
	return
}

type Handler interface {
	ServeProc(ctx context.Context, p *proc.P) error
}

type HandlerFunc func(context.Context, *proc.P) error

func (handle HandlerFunc) ServeProc(ctx context.Context, p *proc.P) error {
	return handle(ctx, p)
}
