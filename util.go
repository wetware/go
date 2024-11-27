package ww

import (
	"context"
	"io"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"go.uber.org/multierr"
)

func NewP2PHostWithMDNS(ctx context.Context, opt ...libp2p.Option) (host.Host, error) {
	// Set up libp2p and peer discovery
	////
	h, err := libp2p.New()
	if err != nil {
		return nil, err
	}

	// Start a multicast DNS service that searches for local
	// peers in the background.
	////
	d, err := MDNS{
		Host:    h,
		Handler: StorePeer{Peerstore: h.Peerstore()},
	}.New()
	if err != nil {
		defer h.Close()
		return nil, err
	}

	return closerHost{Host: h, Closer: d}, nil
}

type closerHost struct {
	host.Host
	io.Closer
}

func (c closerHost) Close() error {
	return multierr.Combine(
		c.Closer.Close(), // first mDNS
		c.Host.Close())   // then the host
}
