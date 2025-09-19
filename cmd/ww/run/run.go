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

	"github.com/ipfs/boxo/path"
	"github.com/libp2p/go-libp2p/core/network"
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
				Value:   "/dns4/localhost/tcp/5001/http",
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

	// Log connection information for async mode
	slog.InfoContext(ctx, "process started in async mode",
		"peer-id", env.Host.ID(),
		"endpoint", p.Endpoint.Name,
		"protocol", string(p.Endpoint.Protocol()),
		"addresses", env.Host.Addrs())

	env.Host.SetStreamHandler(p.Endpoint.Protocol(), func(s network.Stream) {
		defer s.Close()
		slog.InfoContext(ctx, "stream connected",
			"peer-id", s.Conn().RemotePeer(),
			"stream-id", s.ID(),
			"endpoint", p.Endpoint.Name)
		if err := p.Poll(ctx, s, nil); err != nil {
			slog.ErrorContext(ctx, "failed to poll process",
				"id", p.ID(),
				"stream", s.ID(),
				"reason", err)
		}
	})
	defer env.Host.RemoveStreamHandler(p.Endpoint.Protocol())

	<-ctx.Done()
	return ctx.Err()
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
