package util

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// NewDHT creates a client-mode DHT seeded with IPFS peers
func (env *IPFSEnv) NewDHT(ctx context.Context, h host.Host) (*dual.DHT, error) {
	known, err := env.IPFS.Swarm().KnownAddrs(ctx)
	if err != nil {
		return nil, err
	}

	var infos []peer.AddrInfo
	for id, addrs := range known {
		infos = append(infos, peer.AddrInfo{
			ID:    id,
			Addrs: addrs,
		})
	}
	rand.Shuffle(len(infos), func(i, j int) {
		infos[i], infos[j] = infos[j], infos[i]
	})

	slog.DebugContext(ctx, "found known peers from IPFS",
		"count", len(infos))

	return dual.New(ctx, h, dual.DHTOption(
		dht.Mode(dht.ModeClient),
		dht.BootstrapPeers(infos...)))
}

// WaitForDHTReady waits for the DHT to be ready by monitoring both WAN and LAN routing tables
//
// Note: The go-libp2p-kad-dht library doesn't provide explicit events for DHT readiness.
// The recommended approach is to monitor the routing table size against the internal
// minRTRefreshThreshold. The dual DHT queries both WAN and LAN in parallel, so we
// monitor both routing tables to ensure the DHT can handle queries effectively.
func WaitForDHTReady(ctx context.Context, dht *dual.DHT) error {
	// minRTRefreshThreshold is the minimum number of peers required in the routing table
	// for the DHT to be considered ready. This matches the internal constant used by
	// go-libp2p-kad-dht (minRTRefreshThreshold = 2).
	const minRTRefreshThreshold = 2

	slog.DebugContext(ctx, "Monitoring DHT routing tables", "wan", "lan")

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
		case <-ticker.C:
		}

		// Check both routing tables periodically
		// The dual DHT queries both WAN and LAN in parallel, so we need both to be ready
		wanSize := dht.WAN.RoutingTable().Size()
		lanSize := dht.LAN.RoutingTable().Size()
		totalSize := wanSize + lanSize

		if totalSize >= minRTRefreshThreshold {
			return nil
		}

		if err := ctx.Err(); err != nil {
			return err
		}
	}
}
