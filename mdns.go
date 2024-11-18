package ww

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type MDNS struct {
	Host    host.Host
	Handler mdns.Notifee
}

// Serve mDNS to discover peers on the local network
func (m MDNS) Serve(ctx context.Context) error {
	d, err := m.New(ctx)
	if err != nil {
		return err
	}
	defer d.Close()

	<-ctx.Done()
	return ctx.Err()
}

func (m MDNS) New(ctx context.Context) (mdns.Service, error) {
	d := mdns.NewMdnsService(m.Host, "ww.local", m.Handler)
	return d, d.Start()
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
