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

	"github.com/ipfs/boxo/files"
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
	bytecode, err := resolveBinary(ctx, binaryPath)
	if err != nil {
		return fmt.Errorf("failed to resolve binary %s: %w", binaryPath, err)
	}

	// Create wazero runtime
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithDebugInfoEnabled(c.Bool("wasm-debug")).
		WithCloseOnContextDone(true))
	defer runtime.Close(ctx)

	p, err := system.ProcConfig{
		Host:      env.Host,
		Runtime:   runtime,
		Bytecode:  bytecode,
		ErrWriter: c.App.ErrWriter,
	}.New(ctx)
	if err != nil && !errors.Is(err, sys.Errno(0)) {
		return err
	}
	defer p.Close(ctx)

	if !p.Config.Async {
		return nil
	}

	env.Host.SetStreamHandler(p.Endpoint.Protocol(), func(s network.Stream) {
		defer s.Close()
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
func resolveBinary(ctx context.Context, path string) ([]byte, error) {
	// Check if it's an IPFS/IPLD/IPNS path
	if isIPFSPath(path) {
		return resolveIPFSPath(ctx, path)
	}

	// Check if it's an absolute path
	if filepath.IsAbs(path) {
		return os.ReadFile(path)
	}

	// Check if it's a relative path (starts with . or /)
	if len(path) > 0 && (path[0] == '.' || path[0] == '/') {
		return os.ReadFile(path)
	}

	// Check if it's in $PATH
	if resolvedPath, err := exec.LookPath(path); err == nil {
		return os.ReadFile(resolvedPath)
	}

	// Try as a relative path in current directory
	if _, err := os.Stat(path); err == nil {
		return os.ReadFile(path)
	}

	return nil, fmt.Errorf("binary not found: %s", path)
}

// isIPFSPath checks if the path is an IPFS/IPLD/IPNS path
func isIPFSPath(path string) bool {
	return len(path) > 5 && (path[:5] == "/ipfs" || path[:5] == "/ipld" || path[:5] == "/ipns")
}

// resolveIPFSPath resolves an IPFS path to WASM bytecode
func resolveIPFSPath(ctx context.Context, pathStr string) ([]byte, error) {
	if env.IPFS == nil {
		return nil, fmt.Errorf("IPFS environment not initialized")
	}

	// Parse the IPFS path
	ipfsPath, err := path.NewPath(pathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid IPFS path: %w", err)
	}

	// Get the file from IPFS
	node, err := env.IPFS.Unixfs().Get(ctx, ipfsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get from IPFS: %w", err)
	}

	// Read the file content
	file, ok := node.(files.File)
	if !ok {
		return nil, fmt.Errorf("IPFS path does not point to a file")
	}

	return io.ReadAll(file)
}

func withEnv(c *cli.Context, config wazero.ModuleConfig) wazero.ModuleConfig {
	for _, s := range c.StringSlice("env") {
		if envvar := strings.SplitN(s, "=", 2); len(envvar[0]) == 0 || len(envvar[1]) == 0 {
			slog.WarnContext(c.Context, "invalid env string",
				"env", s)
			continue
		} else {
			config = config.WithEnv(envvar[0], envvar[1])
		}
	}
	return config
}
