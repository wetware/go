package boot

import (
	"context"
	"log/slog"
	"time"

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
	d, err := m.New()
	if err != nil {
		return err
	}
	defer d.Close()

	<-ctx.Done()
	return ctx.Err()
}

func (m MDNS) New() (mdns.Service, error) {
	d := mdns.NewMdnsService(m.Host, "ww.local", m.Handler)
	return d, d.Start()
}

type PeerHandler struct {
	Peerstore peerstore.Peerstore
	Bootstrapper
}

func (h PeerHandler) HandlePeerFound(info peer.AddrInfo) {
	log := slog.With(
		"peer", info.ID,
		"addrs", info.Addrs)

	addrs, err := peer.AddrInfoToP2pAddrs(&info)
	if err != nil {
		log.Error("failed to parse discovered p2p addr",
			"reason", err)
	}

	for _, a := range addrs {
		h.Peerstore.SetAddr(info.ID, a, peerstore.AddressTTL)
	}

	// Bootstrap the DHT
	const timeout = time.Second * 5
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := h.Bootstrap(ctx); err != nil {
		log.Error("failed to bootstrap dht",
			"reason", err)
	}
}
