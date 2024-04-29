package boot

import (
	"log/slog"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
)

type EvtPeerFound struct {
	Peer peer.AddrInfo
}

type EmitPeerFound struct {
	event.Emitter
}

func (e EmitPeerFound) HandlePeerFound(info peer.AddrInfo) {
	event := EvtPeerFound{
		Peer: info,
	}

	if err := e.Emit(event); err != nil {
		slog.Error("failed to emit event",
			"reason", err,
			"event", event)
	}
}
