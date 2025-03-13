package cluster_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/cluster"
	"github.com/wetware/go/system"
)

func TestHostTransport(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create two libp2p hosts
	h1, err := libp2p.New()
	require.NoError(t, err)
	defer h1.Close()

	h2, err := libp2p.New()
	require.NoError(t, err)
	defer h2.Close()

	// Create transports
	t1 := cluster.NewHostTransport(ctx, h1)
	t2 := cluster.NewHostTransport(ctx, h2)

	// Connect the hosts
	err = connectHosts(h1, h2)
	require.NoError(t, err)

	// Verify connection and protocol support
	require.Eventually(t, func() bool {
		return h1.Network().Connectedness(h2.ID()) == network.Connected &&
			h2.Network().Connectedness(h1.ID()) == network.Connected
	}, time.Second, 10*time.Millisecond, "hosts failed to connect")

	t.Run("SendTo and receive packet", func(t *testing.T) {
		payload := []byte("hello world")
		addr := h2.Addrs()[0].String() + "/p2p/" + h2.ID().String()

		// Start receiving before sending
		done := make(chan struct{})
		go func() {
			defer close(done)
			select {
			case packet := <-t2.Packets():
				require.Equal(t, payload, packet.Buf)
				require.Equal(t, h1.ID().String(), packet.From.String())
			case <-time.After(time.Second):
				t.Error("timeout waiting for packet")
			}
		}()

		// Send packet from t1 to t2
		err := t1.SendTo(payload, addr)
		require.NoError(t, err)

		// Wait for receive goroutine
		<-done
	})

	t.Run("SendTo error cases", func(t *testing.T) {
		t.Run("invalid multiaddr", func(t *testing.T) {
			err := t1.SendTo([]byte("test"), "not-a-multiaddr")
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid address")
		})

		t.Run("valid multiaddr with invalid peer", func(t *testing.T) {
			pid := "12D3KooWD3eckifWpRn9wQpMG9R9hX3sD158z7EqHWmweQAJU5SA"
			addr := fmt.Sprintf("/ip4/127.0.0.1/tcp/1234/p2p/%s", pid)

			err := t1.SendTo([]byte("test"), addr)
			require.Error(t, err)
			require.Contains(t, err.Error(), "failed to open stream")
		})
	})

	t.Run("DialAndConnect and Stream", func(t *testing.T) {
		addr := h2.Addrs()[0].String() + "/p2p/" + h2.ID().String()

		// Start accepting connections
		accepted := make(chan struct{})
		go func() {
			defer close(accepted)
			select {
			case conn := <-t2.Stream():
				require.NotNil(t, conn)
				require.Equal(t, h1.ID().String(), conn.RemoteAddr().String())

				// Test bidirectional communication
				msg := []byte("hello from server")
				n, err := conn.Write(msg)
				require.NoError(t, err)
				require.Equal(t, len(msg), n)

				buf := make([]byte, len(msg))
				n, err = io.ReadFull(conn, buf)
				require.NoError(t, err)
				require.Equal(t, len(msg), n)
				require.Equal(t, msg, buf)

				conn.Close()
			case <-time.After(time.Second):
				t.Error("timeout waiting for connection")
			}
		}()

		// Dial connection
		conn, err := t1.DialAndConnect(addr, 500*time.Millisecond)
		require.NoError(t, err)
		defer conn.Close()

		// Test bidirectional communication
		msg := make([]byte, 17) // "hello from server" is 17 bytes
		n, err := io.ReadFull(conn, msg)
		require.NoError(t, err)
		require.Equal(t, 17, n)

		n, err = conn.Write(msg)
		require.NoError(t, err)
		require.Equal(t, len(msg), n)

		// Wait for accept goroutine
		select {
		case <-accepted:
		case <-time.After(time.Second):
			t.Fatal("failed to accept connection within timeout")
		}
	})

	t.Run("Shutdown", func(t *testing.T) {
		err := t1.Shutdown()
		require.NoError(t, err)

		// Verify channels are closed
		_, ok := <-t1.Packets()
		require.False(t, ok, "packets channel should be closed")

		_, ok = <-t1.Stream()
		require.False(t, ok, "stream channel should be closed")

		// Verify can't send packets
		addr := h2.Addrs()[0].String() + "/p2p/" + h2.ID().String()
		err = t2.SendTo([]byte("test"), addr)
		require.Error(t, err, "should not be able to send to shutdown transport")
	})
}

func TestHostAddr(t *testing.T) {
	t.Parallel()

	testID, err := peer.Decode("12D3KooWD3eckifWpRn9wQpMG9R9hX3sD158z7EqHWmweQAJU5SA")
	require.NoError(t, err)

	addr := &cluster.HostAddr{ID: testID}

	t.Run("Network", func(t *testing.T) {
		network := addr.Network()
		require.Equal(t, system.Proto.String(), network)
	})

	t.Run("String", func(t *testing.T) {
		str := addr.String()
		require.Equal(t, testID.String(), str)
	})
}

// Helper function to connect two libp2p hosts
func connectHosts(h1, h2 host.Host) error {
	h2Info := peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	}
	return h1.Connect(context.Background(), h2Info)
}
