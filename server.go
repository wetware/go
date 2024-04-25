package ww

import (
	"context"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type Server struct {
	Host     host.Host
	Behavior NetworkBehavior
}

func (s Server) Serve(ctx context.Context) error {
	// s.Host.SetStreamHandler(...)
	// defer s.Host.RemoveStreamHandler(...)

	ms := mdns.NewMdnsService(s.Host, "", mdnsNotifiee{s.Host})
	if err := ms.Start(); err != nil {
		return err
	}
	defer ms.Close()

	sub, err := s.Host.EventBus().Subscribe([]any{
		new(event.EvtLocalAddressesUpdated),
		new(event.EvtLocalProtocolsUpdated),
		new(event.EvtLocalReachabilityChanged),
		new(event.EvtNATDeviceTypeChanged),
		new(event.EvtPeerConnectednessChanged),
		new(event.EvtPeerIdentificationCompleted),
		new(event.EvtPeerIdentificationFailed),
		new(event.EvtPeerProtocolsUpdated),
	})
	if err != nil {
		return err
	}
	defer sub.Close()

	for {
		select {
		case v := <-sub.Out():
			switch e := v.(type) {
			case event.EvtLocalAddressesUpdated:
				s.Behavior.OnLocalAddrsUpdated(ctx, e)

			case event.EvtLocalProtocolsUpdated,
				event.EvtLocalReachabilityChanged,
				event.EvtNATDeviceTypeChanged,
				event.EvtPeerConnectednessChanged,
				event.EvtPeerIdentificationCompleted,
				event.EvtPeerIdentificationFailed,
				event.EvtPeerProtocolsUpdated:
				// ignore

			default:
				panic(UnhandledEvent{e})
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type mdnsNotifiee struct {
	Host host.Host
}

func (n mdnsNotifiee) HandlePeerFound(info peer.AddrInfo) {
	defer slog.Debug("found peer",
		"peer", info.ID)

	n.Host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.AddressTTL)
}
