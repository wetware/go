package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ipfs/boxo/path"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/experimental/sys"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/cmd/internal/flags"
	"github.com/wetware/go/system"
)

var env Env

func Command() *cli.Command {
	return &cli.Command{
		// ww run <binary> [args...]
		////
		Name:      "run",
		ArgsUsage: "<binary> [args...]",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Value:   "/ip4/127.0.0.1/tcp/5001/http",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				EnvVars: []string{"WW_PORT"},
				Value:   2020,
			},
			&cli.BoolFlag{
				Name:    "wasm-debug",
				Usage:   "enable wasm debug info",
				EnvVars: []string{"WW_WASM_DEBUG"},
			},
			&cli.BoolFlag{
				Name:    "async",
				Usage:   "run in async mode for stream processing",
				EnvVars: []string{"WW_ASYNC"},
			},
		}, flags.CapabilityFlags()...),

		// Environment hooks.
		////
		Before: func(c *cli.Context) (err error) {
			env, err = EnvConfig{
				IPFS: c.String("ipfs"),
				Port: c.Int("port"),
			}.New(c.Context)
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

func Main(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	// Get the binary path from arguments
	binaryPath := c.Args().First()
	if binaryPath == "" {
		return fmt.Errorf("no binary specified")
	}

	// Resolve the binary path to WASM bytecode
	f, err := resolveBinary(ctx, binaryPath)
	if err != nil {
		return fmt.Errorf("failed to resolve binary %s: %w", binaryPath, err)
	}
	defer f.Close()

	// Create wazero runtime
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithDebugInfoEnabled(c.Bool("wasm-debug")).
		WithCloseOnContextDone(true))
	defer runtime.Close(ctx)

	p, err := system.ProcConfig{
		Host:      env.Host,
		Runtime:   runtime,
		Src:       f,
		Env:       c.StringSlice("env"),
		Args:      c.Args().Slice(),
		ErrWriter: c.App.ErrWriter,
		Async:     c.Bool("async"),
	}.New(ctx)
	if err != nil && !errors.Is(err, sys.Errno(0)) {
		return err
	}
	defer p.Close(ctx)

	if !p.Config.Async {
		return nil
	}

	sub, err := env.Host.EventBus().Subscribe([]any{
		new(event.EvtPeerIdentificationCompleted),
		new(event.EvtPeerIdentificationFailed),
		new(event.EvtAutoRelayAddrsUpdated),
		new(event.EvtHostReachableAddrsChanged),
		new(event.EvtLocalReachabilityChanged),
		new(event.EvtPeerProtocolsUpdated),
		new(event.EvtLocalProtocolsUpdated),
		new(event.EvtPeerConnectednessChanged),
		new(event.EvtNATDeviceTypeChanged),
		new(event.EvtLocalAddressesUpdated)})
	if err != nil {
		return fmt.Errorf("failed to subscribe to event loop: %w", err)
	}
	defer sub.Close()

	// Log connection information for async mode
	slog.InfoContext(ctx, "process started in async mode",
		"peer", env.Host.ID(),
		"endpoint", p.Endpoint.Name)

	// Set up stream handler that matches both exact protocol and with method suffix
	baseProto := p.Endpoint.Protocol()
	env.Host.SetStreamHandlerMatch(baseProto, func(protocol protocol.ID) bool {
		// Match exact base protocol (/ww/0.1.0/<proc-id>) or with method suffix (/ww/0.1.0/<proc-id>/<method>)
		return protocol == baseProto || strings.HasPrefix(string(protocol), string(baseProto)+"/")
	}, func(s network.Stream) {
		defer s.CloseRead()

		// Extract method from protocol string
		method := "poll" // default
		protocolStr := string(s.Protocol())
		if strings.HasPrefix(protocolStr, string(baseProto)+"/") {
			// Extract method from /ww/0.1.0/<proc-id>/<method>
			parts := strings.Split(protocolStr, "/")
			if len(parts) > 0 {
				method = parts[len(parts)-1]
			}
		}

		slog.InfoContext(ctx, "stream connected",
			"peer", s.Conn().RemotePeer(),
			"stream-id", s.ID(),
			"endpoint", p.Endpoint.Name,
			"method", method)
		if err := p.ProcessMessage(ctx, s, method); err != nil {
			slog.ErrorContext(ctx, "failed to poll process",
				"id", p.ID(),
				"stream", s.ID(),
				"method", method,
				"reason", err)
		}
	})
	defer env.Host.RemoveStreamHandler(baseProto)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v := <-sub.Out():
			switch ev := v.(type) {
			case event.EvtNATDeviceTypeChanged:
				slog.InfoContext(ctx, "NAT device type changed",
					"device type", ev.NatDeviceType)
			case event.EvtLocalReachabilityChanged:
				slog.InfoContext(ctx, "local reachability changed",
					"reachability", ev.Reachability)
			case event.EvtHostReachableAddrsChanged:
				slog.DebugContext(ctx, "host reachable addresses changed",
					"reachable", ev.Reachable,
					"unreachable", ev.Unreachable,
					"unknown", ev.Unknown)
			case event.EvtLocalAddressesUpdated:
				r, err := ev.SignedPeerRecord.Record()
				if err != nil {
					return err
				}
				signer, err := peer.IDFromPublicKey(ev.SignedPeerRecord.PublicKey)
				if err != nil {
					return err
				}
				slog.DebugContext(ctx, "local addresses updated",
					"current", ev.Current,
					"diffs", ev.Diffs,
					"peer", r.(*peer.PeerRecord).PeerID,
					"addrs", r.(*peer.PeerRecord).Addrs,
					"seq", r.(*peer.PeerRecord).Seq,
					"signer", signer)
			case event.EvtPeerIdentificationCompleted:
				slog.DebugContext(ctx, "peer identification completed",
					"peer", ev.Peer,
					"agent-version", ev.AgentVersion,
					"protocol-version", ev.ProtocolVersion,
					"protocols", ev.Protocols)
			case event.EvtPeerIdentificationFailed:
				slog.WarnContext(ctx, "peer identification failed",
					"peer", ev.Peer,
					"reason", ev.Reason)
			case event.EvtAutoRelayAddrsUpdated:
				slog.DebugContext(ctx, "auto relay addresses updated",
					"addresses", ev.RelayAddrs)
			case event.EvtPeerProtocolsUpdated:
				slog.DebugContext(ctx, "peer protocols updated",
					"peer", ev.Peer,
					"added", ev.Added,
					"removed", ev.Removed)
			case event.EvtLocalProtocolsUpdated:
				slog.DebugContext(ctx, "local protocols updated",
					"added", ev.Added,
					"removed", ev.Removed)
			case event.EvtPeerConnectednessChanged:
				slog.DebugContext(ctx, "peer connectedness changed",
					"peer", ev.Peer,
					"connectedness", ev.Connectedness)

			default:
				panic(v) // unhandled event
			}
		}
	}
}

// resolveBinary resolves a binary path to WASM bytecode
func resolveBinary(ctx context.Context, name string) (io.ReadCloser, error) {
	// Parse the IPFS path
	ipfsPath, err := path.NewPath(name)
	if err == nil {
		return env.LoadIPFSFile(ctx, ipfsPath)
	}

	// Check if it's an absolute path
	if filepath.IsAbs(name) {
		return os.Open(name)
	}

	// Check if it's a relative path (starts with . or /)
	if len(name) > 0 && (name[0] == '.' || name[0] == '/') {
		return os.Open(name)
	}

	// Check if it's in $PATH
	if resolvedPath, err := exec.LookPath(name); err == nil {
		return os.Open(resolvedPath)
	}

	// Try as a relative path in current directory
	if _, err := os.Stat(name); err == nil {
		return os.Open(name)
	}

	return nil, fmt.Errorf("binary not found: %s", name)
}
