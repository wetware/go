package boot

import (
	"context"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/lthibault/jitterbug/v2"
	"github.com/wetware/go/system"
)

// DHT implements a peer discovery service using the dual DHT.
type DHT struct {
	Env *system.Env
}

func (d DHT) String() string {
	return "dht"
}

// Serve starts the DHT service and begins discovering peers.
// It creates a RoutingDiscovery from the dual DHT and continuously
// searches for peers in the given namespace.
func (d DHT) Serve(ctx context.Context) error {
	// Create a RoutingDiscovery from the dual DHT
	disc := routing.NewRoutingDiscovery(d.Env.DHT)

	// Get the namespace from the protocol path
	ns := system.Proto.String()

	// Start discovering peers
	d.Env.Log().DebugContext(ctx, "service started",
		"service", d.String(),
		"namespace", ns)

	jitter := jitterbug.Uniform{
		Source: rand.New(rand.NewSource(time.Now().UnixNano())),
		Min:    peerstore.AddressTTL / 2,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Advertise ourselves in the namespace with 1 hour TTL
		ttl, err := disc.Advertise(ctx, ns, discovery.TTL(peerstore.AddressTTL))
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			d.Env.Log().WarnContext(ctx, "failed to advertise in DHT",
				"error", err,
				"service", d.String())
			continue
		}

		d.Env.Log().DebugContext(ctx, "advertised in DHT",
			"service", d.String(),
			"namespace", ns,
			"ttl", ttl)

		// Find peers in the namespace
		peerChan, err := disc.FindPeers(ctx, ns)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			d.Env.Log().WarnContext(ctx, "failed to find peers",
				"error", err,
				"service", d.String())
			continue
		}

		// Handle discovered peers
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case peer, ok := <-peerChan:
				if !ok {
					goto NEXT_ITERATION
				}
				if peer.ID != d.Env.Host.ID() { // skip self
					d.Env.HandlePeerFound(peer)
				}
			}
		}

	NEXT_ITERATION:
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(jitter.Jitter(ttl)):
		}
	}
}
