package cat

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mr-tron/base58"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/util"
)

var env util.IPFSEnv

func Command() *cli.Command {
	return &cli.Command{
		Name:      "cat",
		Usage:     "connect stdin/stdout to a remote peer's stream via IPFS",
		ArgsUsage: "<peer-id> <endpoint>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Value:   "/dns4/localhost/tcp/5001/http",
			},
		},

		// Environment hooks
		Before: func(c *cli.Context) (err error) {
			err = env.Boot(c.String("ipfs"))
			return
		},
		After: func(c *cli.Context) (err error) {
			err = env.Close()
			return
		},

		// Main
		////
		Action: Main,
	}
}

func Main(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	// Get arguments
	args := c.Args().Slice()
	if len(args) != 2 {
		return fmt.Errorf("usage: ww cat <peer-id> <endpoint>")
	}

	peerIDStr := args[0]
	endpointStr := args[1]

	// Parse peer ID
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer ID %s: %w", peerIDStr, err)
	}

	// Validate base58 endpoint (but use it as-is)
	_, err = base58.FastBase58Decoding(endpointStr)
	if err != nil {
		return fmt.Errorf("invalid endpoint %s: %w", endpointStr, err)
	}

	// Create the full protocol ID
	protocolID := protocol.ID("/ww/0.1.0/" + endpointStr)

	slog.InfoContext(ctx, "connecting to peer via IPFS",
		"peer", peerID,
		"endpoint", endpointStr,
		"protocol", string(protocolID))

	// Create a minimal libp2p host for client-only connection
	host, err := libp2p.New(
		libp2p.NoListenAddrs, // Don't listen, just connect
	)
	if err != nil {
		return fmt.Errorf("failed to create libp2p host: %w", err)
	}
	defer host.Close()

	slog.InfoContext(ctx, "created client host", "peer-id", host.ID())

	// For now, we'll assume the remote peer is listening on localhost:2020
	// In a real scenario, this would come from IPFS peer discovery or manual configuration
	remoteAddr := "/ip4/127.0.0.1/tcp/2020"

	slog.InfoContext(ctx, "connecting to remote peer", "address", remoteAddr)

	// Parse the multiaddr and connect
	addrInfo, err := peer.AddrInfoFromString(remoteAddr + "/p2p/" + peerIDStr)
	if err != nil {
		return fmt.Errorf("invalid peer address: %w", err)
	}

	// Connect to the peer
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := host.Connect(ctx, *addrInfo); err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	slog.InfoContext(ctx, "connected to remote peer")

	// Open a stream to the remote peer
	stream, err := host.NewStream(ctx, peerID, protocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	slog.InfoContext(ctx, "stream opened successfully", "stream-id", stream.ID())

	// Set up bidirectional copying
	done := make(chan error, 2)

	// Copy from local stdin to remote stream
	go func() {
		defer stream.CloseWrite()
		_, err := io.Copy(stream, os.Stdin)
		done <- err
	}()

	// Copy from remote stream to local stdout
	go func() {
		defer stream.CloseRead()
		_, err := io.Copy(os.Stdout, stream)
		done <- err
	}()

	// Wait for either direction to complete
	select {
	case err := <-done:
		if err != nil {
			slog.ErrorContext(ctx, "stream error", "error", err)
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
