package system

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	core_routing "github.com/libp2p/go-libp2p/core/routing"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/wetware/go/util"
)

type Env struct {
	NS   string // fully qualified namespace
	IPFS iface.CoreAPI
	Host host.Host
	DHT  core_routing.Routing
}

func (env Env) Log() *slog.Logger {
	return slog.With("peer", env.Host.ID())
}

// HandlePeerFound is called when a peer is discovered, such as when
// using MDNS or DHT discovery.
func (env Env) HandlePeerFound(info peer.AddrInfo) {
	pstore := env.Host.Peerstore()
	pstore.AddAddrs(info.ID, info.Addrs, peerstore.AddressTTL)
	env.Log().Info("peer discovered", "found", info.ID)

	// 5s delay to bootstrap dht, which is an asynchronous operation,
	// is plenty.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := env.DHT.Bootstrap(ctx); err != nil {
		slog.Warn("bootstrap failed",
			"reason", err)
	}
}

func (env Env) Load(ctx context.Context, p string) ([]byte, error) {
	path, err := path.NewPath(p)
	if err != nil {
		return nil, fmt.Errorf("invalid IPFS path %q: %w", p, err)
	}

	node, err := env.IPFS.Unixfs().Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer node.Close()

	return util.LoadByteCode(ctx, node)
}

func (env Env) NewUnixFS(ctx context.Context) UnixFS {
	return UnixFS{
		Ctx:  ctx,
		Unix: env.IPFS.Unixfs(),
	}
}

func (env Env) FilterPeers(selector func(ma.Multiaddr) bool) (ps []peer.AddrInfo) {
	for _, pid := range env.Host.Peerstore().Peers() {
		var ms []ma.Multiaddr
		for _, addr := range env.Host.Peerstore().Addrs(pid) {
			if selector(addr) {
				ms = append(ms, addr)
			}
		}

		if len(ms) > 0 {
			ps = append(ps, peer.AddrInfo{
				ID:    pid,
				Addrs: ms,
			})
		}
	}

	return
}

// PublicBootstrapPeers filters out peers that don't have public IP addresses.
// It returns a slice of peer.AddrInfo for peers that have public IP addresses.
func (env Env) PublicBootstrapPeers() []peer.AddrInfo {
	return env.FilterPeers(func(m ma.Multiaddr) bool {
		ip, err := extractIPFromMultiaddr(m)
		return err != nil && ip != nil && !isPrivateIP(ip) // public ip or relay
	})
}

// PrivateBootstrapPeers filters out peers that don't have private IP addresses.
// It returns a slice of peer.AddrInfo for peers that have RFC1918 private IPs.
func (env Env) PrivateBootstrapPeers() []peer.AddrInfo {
	return env.FilterPeers(func(m ma.Multiaddr) bool {
		ip, err := extractIPFromMultiaddr(m)
		return err == nil && ip != nil && isPrivateIP(ip) // private ip
	})
}

// https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.40.0/p2p/net/swarm#DefaultDialRanker
// https://www.rfc-editor.org/rfc/rfc1918.html
var privateRanges = []net.IPNet{
	{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},     // 10.0.0.0/8
	{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},  // 172.16.0.0/12
	{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)}, // 192.168.0.0/16
}

// isPrivateIP checks if an IP is within RFC1918 private ranges.
func isPrivateIP(ip net.IP) bool {
	return inRange(ip, privateRanges)
}

// inRange checks if an IP is within any of the provided IP ranges.
func inRange(ip net.IP, ipRanges []net.IPNet) bool {
	for _, ipRange := range ipRanges {
		if ipRange.Contains(ip) {
			return true
		}
	}
	return false
}

// extractIPFromMultiaddr tries to extract an IPv4 address from a multiaddr.
func extractIPFromMultiaddr(addr ma.Multiaddr) (net.IP, error) {
	ipStr, err := addr.ValueForProtocol(ma.P_IP4)
	if err != nil {
		return nil, err
	}
	return net.ParseIP(ipStr), nil
}
