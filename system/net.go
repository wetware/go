package system

import (
	"context"

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

	n.Host.SetStreamHandlerMatch(
		handler.Proto(),
		handler.Match,
		handler.Bind(ctx))
	return func() { n.Host.RemoveStreamHandler(handler.Proto()) }
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
