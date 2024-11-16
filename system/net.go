package system

import (
	"context"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
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

// StorePeer is a peer handler that inserts the peer in the
// supplied Peerstore.
type StorePeer struct {
	peerstore.Peerstore
}

func (s StorePeer) HandlePeerFound(info peer.AddrInfo) {
	for _, addr := range info.Addrs {
		s.AddAddr(info.ID, addr, peerstore.AddressTTL) // assume a dynamic environment
	}
}

type Handler interface {
	ServeProc(ctx context.Context, p *proc.P) error
}

type HandlerFunc func(context.Context, *proc.P) error

func (handle HandlerFunc) ServeProc(ctx context.Context, p *proc.P) error {
	return handle(ctx, p)
}
