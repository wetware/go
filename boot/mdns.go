package boot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type MDNS struct {
	Host        host.Host
	ServiceName string
}

func (m MDNS) String() string {
	if m.ServiceName == "" {
		m.ServiceName = mdns.ServiceName
	}

	return fmt.Sprintf("MDNS{%s}", m.ServiceName)
}

func (m *MDNS) Serve(ctx context.Context) error {
	// Set up mDNS for local network discovery
	e, err := m.Host.EventBus().Emitter(new(EvtPeerFound))
	if err != nil {
		return err
	}
	defer e.Close()

	ms := mdns.NewMdnsService(m.Host, m.ServiceName, EmitPeerFound{
		Emitter: e,
	})
	defer ms.Close() // idempotent

	if err := ms.Start(); err != nil {
		return err
	}

	slog.DebugContext(ctx, "mdns started")
	defer slog.DebugContext(ctx, "mdns stopped")

	<-ctx.Done()

	if err := ms.Close(); err != nil {
		return err
	}

	return ctx.Err()
}
