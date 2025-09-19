package cat

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mr-tron/base58"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/cmd/internal/flags"
	"github.com/wetware/go/cmd/ww/run"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "cat",
		Usage:     "connect stdin/stdout to a remote peer's stream",
		ArgsUsage: "<peer-id> <endpoint>",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Value:   "/dns4/localhost/tcp/5001/http",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				EnvVars: []string{"WW_PORT"},
				Value:   2020,
			},
		}, flags.CapabilityFlags()...),

		// Environment hooks.
		////
		Before: func(c *cli.Context) (err error) {
			// Initialize environment similar to run command
			env, err = run.EnvConfig{
				NS:   c.String("ns"),
				IPFS: c.String("ipfs"),
				Port: c.Int("port"),
				MDNS: c.Bool("mdns"),
			}.New()
			return
		},
		After: func(c *cli.Context) error {
			return env.Close()
		},

		// Main
		////
		Action: Main,
	}
}

var env run.Env

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

	// Decode base58 endpoint
	endpointBytes, err := base58.FastBase58Decoding(endpointStr)
	if err != nil {
		return fmt.Errorf("invalid endpoint %s: %w", endpointStr, err)
	}

	// Create the full protocol ID
	protocolID := protocol.ID("/ww/0.1.0/" + string(endpointBytes))

	slog.InfoContext(ctx, "connecting to peer",
		"peer", peerID,
		"endpoint", endpointStr,
		"protocol", string(protocolID))

	// Open a stream to the remote peer
	stream, err := env.Host.NewStream(ctx, peerID, protocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	slog.InfoContext(ctx, "stream opened", "stream", stream.ID())

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
