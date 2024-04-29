package boot

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/thejerf/suture/v4"
)

type StaticPeers struct {
	Host  host.Host
	Addrs []string
}

func (static StaticPeers) String() string {
	return fmt.Sprintf("StaticPeers{len=%d}", len(static.Addrs))
}

func (static StaticPeers) Serve(ctx context.Context) error {
	as, err := static.Parse()
	if err != nil {
		return err
	}

	peers, err := peer.AddrInfosFromP2pAddrs(as...)
	if err != nil {
		return err
	}

	e, err := static.Host.EventBus().Emitter(new(EvtPeerFound))
	if err != nil {
		return err
	}

	for _, info := range peers {
		if static.Host.ID() == info.ID {
			continue
		}

		EmitPeerFound{
			Emitter: e,
		}.HandlePeerFound(info)

	}

	return suture.ErrDoNotRestart
}

func (static StaticPeers) Parse() ([]ma.Multiaddr, error) {
	var ms []ma.Multiaddr
	for _, s := range static.Addrs {
		m, err := ma.NewMultiaddr(s)
		if err != nil {
			return nil, err
		}

		ms = append(ms, m)
	}

	return ms, nil
}
