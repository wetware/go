package ww

import (
	"context"
	"log/slog"

	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/event"
)

const Proto = "/ww/0.0.0"

type NetworkBehavior interface {
	OnLocalAddrsUpdated(context.Context, event.EvtLocalAddressesUpdated)
}

type DefaultBehavior struct {
	Public iface.CoreAPI
}

func (b *DefaultBehavior) OnLocalAddrsUpdated(
	ctx context.Context,
	e event.EvtLocalAddressesUpdated,
) {
	slog.DebugContext(ctx, "local addrs updated",
		"addrs", e.Current)
}
