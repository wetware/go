package system

import (
	"context"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

type Env struct {
	NS   string // fully qualified namespace
	IPFS iface.CoreAPI
	Host host.Host
	DHT  interface {
		Bootstrap(context.Context) error
		Provide(context.Context, cid.Cid, bool) error
		FindPeer(context.Context, peer.ID) (peer.AddrInfo, error)
	}
}

func (env Env) Log() *slog.Logger {
	return slog.With("peer", env.Host.ID())
}

func (env Env) HandlePeerFound(info peer.AddrInfo) {
	// TODO:  do we want to move this to boot/mdns.go?   Currently, this
	// callback is used exclusively by the MDNS discovery system, but it
	// can be used by other discovery systems in principle.

	pstore := env.Host.Peerstore()
	pstore.AddAddrs(info.ID, info.Addrs, peerstore.AddressTTL)
	env.Log().Info("peer discovered", "found", info.ID)

	// 5s delay to bootstrap dht, which is an asynchronous operation,
	// is plenty.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := env.Host.Connect(ctx, peer.AddrInfo{
		ID:    info.ID,
		Addrs: info.Addrs,
	}); err != nil {
		env.Log().Debug("failed to connect to peer",
			"reason", err,
			"peer", info.ID,
			"addrs", info.Addrs)
	} else if err := env.DHT.Bootstrap(ctx); err != nil {
		env.Log().Error("failed to bootstrap dht",
			"reason", err)
	}
}

func (env Env) Load(ctx context.Context, p string) ([]byte, error) {
	path, err := path.NewPath(p)
	if err != nil {
		return nil, err
	}

	node, err := env.IPFS.Unixfs().Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer node.Close()

	// TODO: improve
	switch n := node.(type) {
	case files.File:
		return io.ReadAll(n)
	case files.Directory:
		entries := n.Entries()
		for entries.Next() {
			if entries.Name() == "main.wasm" {
				return io.ReadAll(entries.Node().(io.Reader))
			}
		}
	}

	return nil, errors.New("not found")
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

func (env Env) PublicBootstrapPeers() []peer.AddrInfo {
	return env.FilterPeers(func(m ma.Multiaddr) bool {
		ip, err := extractIPFromMultiaddr(m)
		return err != nil && ip != nil && !isPrivateIP(ip) // public ip or relay
	})
}

// PrivateBootstrapPeers filters out peers that don't have private IP addresses.
// It returns a slice of peer.AddrInfo for peers that have RFC1918 private IPs
// and are currently connected to the local host.
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
