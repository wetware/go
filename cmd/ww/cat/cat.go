package cat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"syscall"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/cmd/internal/flags"
	"github.com/wetware/go/util"
)

var env util.IPFSEnv

func Command() *cli.Command {
	return &cli.Command{
		Name:      "cat",
		ArgsUsage: "<peer> <proc> [method]",
		Usage:     "Connect to a peer and execute a procedure over a stream",
		Description: `Connect to a specified peer and execute a procedure over a custom protocol stream.
The command will:
1. Initialize IPFS environment for stream forwarding
2. Use IPFS to establish connection to the specified peer
3. Forward the stream using the /ww/0.1.0/<proc> protocol
4. Bind the stream to stdin/stdout for communication

Examples:
  ww cat QmPeer123 /echo
  ww cat 12D3KooW... /myproc echo
  ww cat 12D3KooW... /myproc poll`,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Value:   "/ip4/127.0.0.1/tcp/5001/http",
				Usage:   "IPFS API endpoint",
			},
		}, append(flags.CapabilityFlags(), flags.P2PFlags()...)...),

		Before: func(c *cli.Context) error {
			return env.Boot(c.String("ipfs"))
		},
		After: func(c *cli.Context) error {
			return env.Close()
		},

		Action: Main,
	}
}

func Main(c *cli.Context) error {
	ctx, cancel := context.WithTimeout(c.Context, c.Duration("timeout"))
	defer cancel()

	if c.NArg() < 3 {
		return cli.Exit("cat requires 2-3 arguments: <peer> <proc> [method]", 1)
	}

	peerIDStr := c.Args().Get(0)
	procName := c.Args().Get(1)
	method := c.Args().Get(2)

	// Parse peer ID
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID %s: %w", peerIDStr, err)
	}

	// Construct protocol ID
	protocolID := protocol.ID("/ww/0.1.0/" + procName)
	if method != "" && method != "poll" {
		protocolID = protocol.ID("/ww/0.1.0/" + procName + "/" + method)
	}

	// Create libp2p host in client mode
	h, err := util.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create host: %w", err)
	}
	defer h.Close()

	dht, err := env.NewDHT(ctx, h)
	if err != nil {
		return fmt.Errorf("failed to create DHT client: %w", err)
	}
	defer dht.Close()

	// Set up DHT readiness monitoring BEFORE bootstrapping
	slog.DebugContext(ctx, "setting up DHT readiness monitoring")
	readyChan := make(chan error, 1)
	go func() {
		readyChan <- util.WaitForDHTReady(ctx, dht)
	}()

	// Bootstrap the DHT to populate routing table with IPFS peers
	slog.DebugContext(ctx, "bootstrapping DHT")
	if err := dht.Bootstrap(ctx); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// Wait for DHT to be ready
	slog.DebugContext(ctx, "waiting for DHT routing table to populate")
	if err := <-readyChan; err != nil {
		slog.WarnContext(ctx, "DHT may not be fully ready", "error", err)
	}

	// Use DHT for peer discovery
	slog.DebugContext(ctx, "searching for peer via DHT", "peer", peerID.String()[:12])

	// Try to find the peer using DHT
	peerInfo, err := dht.FindPeer(ctx, peerID)
	if err != nil {
		return fmt.Errorf("failed to find peer %s via DHT: %w", peerID, err)
	}

	slog.DebugContext(ctx, "target peer found via DHT", "peer", peerInfo.ID.String()[:12])
	if err := h.Connect(ctx, peerInfo); err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", peerInfo.ID, err)
	}

	// Open stream to peer
	stream, err := h.NewStream(ctx, peerID, protocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream to peer %s: %w", peerID, err)
	}
	defer stream.Close()

	// Display connection banner
	fmt.Printf("⚗️  Wetware Stream Connected\n")
	fmt.Printf("   Peer: %s...\n", peerID.String()[:12])
	fmt.Printf("   Endpoint: %s\n", protocolID)
	fmt.Printf("   Ctrl+C to exit\n\n")

	// Bind stream to stdin/stdout
	return bindStreamToStdio(ctx, stream)
}

func bindStreamToStdio(ctx context.Context, stream network.Stream) error {
	// Copy data between stream and stdin/stdout
	readDone := make(chan error, 1)
	writeDone := make(chan error, 1)

	// Copy from stream to stdout
	go func() {
		_, err := io.Copy(os.Stdout, stream)
		readDone <- err
	}()

	// Copy from stdin to stream
	go func() {
		_, err := io.Copy(stream, os.Stdin)
		writeDone <- err
	}()

	// Wait for stdin to close (Ctrl+D)
	err := <-writeDone
	if err != nil && err != io.EOF {
		// Check if it's a broken pipe error (expected when stdin closes)
		if !errors.Is(err, syscall.EPIPE) {
			return err
		}
	}

	// Close the write end to signal EOF to remote peer
	// This will trigger the echo server to respond and then close
	stream.CloseWrite()

	// Wait for the remote peer to finish processing and send response
	// The echo server should now process the input and send back the echoed text
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-readDone:
		// Check if this is a graceful closure (EOF) or an error
		if err == nil || errors.Is(err, io.EOF) {
			// Graceful closure - this is expected
			return nil
		}
		// Any other error should be reported
		return err
	}
}
