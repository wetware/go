package cat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/cmd/internal/flags"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "cat",
		ArgsUsage: "<peer> <proc>",
		Usage:     "Connect to a peer and execute a procedure over a stream",
		Description: `Connect to a specified peer and execute a procedure over a custom protocol stream.
The command will:
1. Initialize IPFS environment for stream forwarding
2. Use IPFS to establish connection to the specified peer
3. Forward the stream using the /ww/0.1.0/<proc> protocol
4. Bind the stream to stdin/stdout for communication

Examples:
  ww cat QmPeer123 /echo
  ww cat 12D3KooW... /myproc`,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Value:   "/dns4/localhost/tcp/5001/http",
				Usage:   "IPFS API endpoint",
			},
			&cli.DurationFlag{
				Name:    "timeout",
				EnvVars: []string{"WW_TIMEOUT"},
				Value:   30 * time.Second,
				Usage:   "Connection timeout",
			},
		}, flags.CapabilityFlags()...),

		Action: Main,
	}
}

func Main(c *cli.Context) error {
	ctx, cancel := context.WithTimeout(c.Context, c.Duration("timeout"))
	defer cancel()

	if c.NArg() != 2 {
		return cli.Exit("cat requires exactly two arguments: <peer> <proc>", 1)
	}

	peerIDStr := c.Args().Get(0)
	procName := c.Args().Get(1)

	// Parse peer ID
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID %s: %w", peerIDStr, err)
	}

	// Construct protocol ID
	protocolID := protocol.ID("/ww/0.1.0/" + procName)

	// Create libp2p host in client mode
	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		return fmt.Errorf("failed to create host: %w", err)
	}
	defer h.Close()

	// Set up DHT in client mode
	dht, err := dual.New(
		ctx,
		h,
		dual.DHTOption(
			dht.Mode(dht.ModeClient), // Client mode
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create DHT: %w", err)
	}
	defer dht.Close()

	// Bootstrap DHT
	if err := dht.Bootstrap(ctx); err != nil {
		slog.WarnContext(ctx, "failed to bootstrap DHT", "error", err)
	}

	// Find peer addresses using DHT
	peerInfo, err := dht.FindPeer(ctx, peerID)
	if err != nil {
		return fmt.Errorf("failed to find peer %s: %w", peerID, err)
	}

	// Connect to peer
	if err := h.Connect(ctx, peerInfo); err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", peerID, err)
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

	// Wait for BOTH streams to finish (linger behavior)
	// Don't exit until both the write and read directions are complete
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
