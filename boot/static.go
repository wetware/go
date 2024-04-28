package boot

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/thejerf/suture/v4"
)

type StaticPeers struct {
	Bus   event.Bus
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

	e, err := static.Bus.Emitter(new(EvtPeerFound))
	if err != nil {
		return err
	}

	for _, info := range peers {
		EmitPeerFound{
			TTL:     peerstore.PermanentAddrTTL,
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
