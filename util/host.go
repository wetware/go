package util

import (
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ws "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	webtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"
)

func NewClient(opts ...libp2p.Option) (host.Host, error) {
	return libp2p.New(append(DefaultClientOptions(), opts...)...)
}

func DefaultClientOptions() []libp2p.Option {
	return []libp2p.Option{
		libp2p.NoListenAddrs,
	}
}

func NewServer(port int, opts ...libp2p.Option) (host.Host, error) {
	return libp2p.New(append(DefaultServerOptions(port), opts...)...)
}

func DefaultServerOptions(port int) []libp2p.Option {
	return []libp2p.Option{
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(quic.NewTransport),
		libp2p.Transport(ws.New),
		libp2p.Transport(webtransport.New),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),
			fmt.Sprintf("/ip6/::/tcp/%d", port),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", port),
			fmt.Sprintf("/ip6/::/udp/%d/quic-v1", port),
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ws", port),
			fmt.Sprintf("/ip6/::/tcp/%d/ws", port),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", port),
			fmt.Sprintf("/ip6/::/udp/%d/quic-v1/webtransport", port),
		),
	}
}
