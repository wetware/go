package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/wetware/go/util"
	"go.uber.org/multierr"
)

func ExpandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

type EnvConfig struct {
	IPFS string
	Port int
}

func (cfg EnvConfig) New(ctx context.Context) (env Env, err error) {
	env.Dir, err = os.MkdirTemp("", "cell-*")
	if err != nil {
		err = fmt.Errorf("failed to create temp directory: %w", err)
		return
	}

	// Initialize IPFS client using embedded IPFSEnv
	////
	if err = env.IPFSEnv.Boot(cfg.IPFS); err != nil {
		err = fmt.Errorf("failed to boot IPFS environment: %w", err)
		return
	}

	// Initialize libp2p host
	////
	env.Host, err = util.NewServer(cfg.Port)
	if err != nil {
		err = fmt.Errorf("failed to create libp2p host: %w", err)
		return
	}

	// Create and bootstrap DHT client
	env.DHT, err = env.NewDHT(ctx, env.Host)
	if err != nil {
		err = fmt.Errorf("failed to create DHT client: %w", err)
		return
	}

	return
}

type Env struct {
	util.IPFSEnv
	Host host.Host
	NS   string
	Dir  string // Temporary directory for cell execution
	DHT  *dual.DHT
}

func (env *Env) Close() error {
	var errors []error

	// Close DHT client
	if env.DHT != nil {
		if err := env.DHT.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close DHT client: %w", err))
		}
	}

	// Close libp2p host
	if env.Host != nil {
		if err := env.Host.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close host: %w", err))
		}
	}

	// Clean up IPFS environment independently
	if err := env.IPFSEnv.Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close IPFS environment: %w", err))
	}

	// Always clean up temporary directory, regardless of other errors
	if env.Dir != "" {
		if err := os.RemoveAll(env.Dir); err != nil {
			errors = append(errors, fmt.Errorf("failed to remove temp directory: %w", err))
		}
	}

	return multierr.Combine(errors...)
}

// ResolveExecPath resolves an executable path, handling both IPFS paths and local filesystem paths.
// For IPFS paths, it uses the import functionality to download and make executable.
// For local paths, it resolves relative paths to absolute paths.
func (env *Env) ResolveExecPath(ctx context.Context, name string) (string, error) {
	// Try to parse as IPFS path first
	if p, err := path.NewPath(name); err == nil {
		// Use import functionality to resolve IPFS paths
		return env.ResolveIPFSPath(ctx, p)
	}

	// Handle non-IPFS paths - resolve relative paths to absolute
	if !filepath.IsAbs(name) {
		absPath, err := filepath.Abs(name)
		if err != nil {
			return "", fmt.Errorf("failed to resolve path %s: %w", name, err)
		}
		name = absPath
	}

	return name, nil
}

// ResolveIPFSPath resolves an IPFS path using the import functionality with caching
func (env *Env) ResolveIPFSPath(ctx context.Context, ipfsPath path.Path) (string, error) {
	// Use the cached import functionality to avoid re-downloading
	importedPath, err := env.IPFSEnv.ImportFromIPFSToDirWithCaching(ctx, ipfsPath, env.Dir, true)
	if err != nil {
		return "", err
	}

	return importedPath, nil
}

// OS returns the operating system name, preferring WW_OS environment variable over runtime.GOOS
func (env *Env) OS() string {
	if os := os.Getenv("WW_OS"); os != "" {
		return os
	}
	return runtime.GOOS
}

// Arch returns the architecture name, preferring WW_ARCH environment variable over runtime.GOARCH
func (env *Env) Arch() string {
	if arch := os.Getenv("WW_ARCH"); arch != "" {
		return arch
	}
	return runtime.GOARCH
}

// LoadIPFSFile resolves an IPFS path to WASM bytecode
func (env *Env) LoadIPFSFile(ctx context.Context, p path.Path) (files.File, error) {
	if env.IPFS == nil {
		return nil, fmt.Errorf("IPFS environment not initialized")
	}

	// Get the file from IPFS
	node, err := env.IPFS.Unixfs().Get(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to get from IPFS: %w", err)
	}

	// Read the file content
	file, ok := node.(files.File)
	if !ok {
		return nil, fmt.Errorf("IPFS path does not point to a file")
	}

	return file, nil
}
