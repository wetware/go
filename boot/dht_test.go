package boot

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/system"
	"golang.org/x/sync/errgroup"
)

// TestDHT verifies that the DHT service correctly handles peer discovery and shutdown.
// It tests:
//  1. Creation and bootstrapping of two DHT nodes
//  2. Connection establishment between peers
//  3. Service startup and peer discovery
//  4. Clean shutdown on context cancellation
//
// The test creates two libp2p hosts with their own DHT instances, connects them,
// and verifies they can discover each other through the DHT service. It then ensures
// the services shut down cleanly when their context is cancelled.
func TestDHT(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create two hosts
	h1, err := libp2p.New()
	require.NoError(t, err)
	defer h1.Close()

	h2, err := libp2p.New()
	require.NoError(t, err)
	defer h2.Close()

	// Create DHT instances for both hosts
	dht1, err := dual.New(ctx, h1)
	require.NoError(t, err)
	defer dht1.Close()

	dht2, err := dual.New(ctx, h2)
	require.NoError(t, err)
	defer dht2.Close()

	// Bootstrap the DHTs
	err = dht1.Bootstrap(ctx)
	require.NoError(t, err)

	err = dht2.Bootstrap(ctx)
	require.NoError(t, err)

	// Create environments for both hosts
	env1 := &system.Env{
		Host: h1,
		DHT:  dht1,
	}

	env2 := &system.Env{
		Host: h2,
		DHT:  dht2,
	}

	// Connect the hosts
	err = connectHosts(h1, h2)
	require.NoError(t, err)

	// Create DHT services
	d1 := &DHT{Env: env1}
	d2 := &DHT{Env: env2}

	// Create a sub-context for the services that we can cancel independently
	svcCtx, svcCancel := context.WithCancel(ctx)
	defer svcCancel() // Ensure services are cancelled even if test fails

	// Start DHT services with errgroup
	g := new(errgroup.Group)

	g.Go(func() error {
		return d1.Serve(svcCtx)
	})

	g.Go(func() error {
		return d2.Serve(svcCtx)
	})

	// Wait for peer discovery
	require.Eventually(t, func() bool {
		return len(h1.Network().Peers()) > 0 && len(h2.Network().Peers()) > 0
	}, 3*time.Second, 100*time.Millisecond, "peers should discover each other")

	// Verify that peers are connected
	require.Contains(t, h1.Network().Peers(), h2.ID(), "h1 should be connected to h2")
	require.Contains(t, h2.Network().Peers(), h1.ID(), "h2 should be connected to h1")

	// Cancel service context
	svcCancel()

	// Wait for services to shut down with timeout
	done := make(chan error, 1)
	go func() {
		done <- g.Wait()
	}()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for services to shut down")
	}
}

// TestDHT_HandlePeerFound verifies that the DHT service correctly handles peer information
// when a new peer is discovered. It tests:
//  1. Proper storage of peer addresses in the peerstore
//  2. Correct handling of multiaddress sets
//  3. Address persistence after peer discovery
//
// The test creates a mock peer with a set of addresses and verifies that when HandlePeerFound
// is called, those addresses are correctly stored in the peerstore. It uses string-based
// comparison of address sets to handle variations in address ordering and representation.
func TestDHT_HandlePeerFound(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a host
	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	// Create DHT instance
	d, err := dual.New(ctx, h)
	require.NoError(t, err)
	defer d.Close()

	// Create environment
	env := &system.Env{
		Host: h,
		DHT:  d,
	}

	// Create a mock peer
	mockPeer, err := libp2p.New()
	require.NoError(t, err)
	defer mockPeer.Close()

	// Create peer info
	peerInfo := peer.AddrInfo{
		ID:    mockPeer.ID(),
		Addrs: mockPeer.Addrs(),
	}

	// Test HandlePeerFound
	env.HandlePeerFound(peerInfo)

	// Verify peer is in peerstore
	storedAddrs := h.Peerstore().Addrs(mockPeer.ID())
	require.NotEmpty(t, storedAddrs, "peer should be added to peerstore")

	// Convert addresses to strings for easier comparison
	expectedAddrs := make(map[string]struct{})
	for _, addr := range peerInfo.Addrs {
		expectedAddrs[addr.String()] = struct{}{}
	}

	storedAddrStrings := make(map[string]struct{})
	for _, addr := range storedAddrs {
		storedAddrStrings[addr.String()] = struct{}{}
	}

	// Compare address sets
	require.Equal(t, expectedAddrs, storedAddrStrings, "stored addresses should match expected addresses")
}

// connectHosts establishes a direct connection between two libp2p hosts.
// It takes the source (h1) and target (h2) hosts and attempts to connect them
// using h2's peer info (ID and addresses). This is a helper function used
// to establish initial connectivity in tests.
func connectHosts(h1, h2 host.Host) error {
	h2Info := peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	}
	return h1.Connect(context.Background(), h2Info)
}
