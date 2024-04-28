package boot

import (
	"log/slog"
	"time"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

type EvtPeerFound struct {
	Peer peer.AddrInfo
	TTL  time.Duration
}

type EmitPeerFound struct {
	TTL time.Duration
	event.Emitter
}

func (e EmitPeerFound) HandlePeerFound(info peer.AddrInfo) {
	if e.TTL <= 0 {
		e.TTL = peerstore.AddressTTL
	}

	event := EvtPeerFound{
		Peer: info,
		TTL:  e.TTL,
	}

	if err := e.Emit(event); err != nil {
		slog.Error("failed to emit event",
			"reason", err,
			"event", event)
	}
}
