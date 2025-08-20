package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
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

type Env struct {
	util.IPFSEnv
	Host host.Host
	Dir  string // Temporary directory for cell execution
}

func (env *Env) Boot(addr string) (err error) {
	// Create temporary directory for cell execution
	env.Dir, err = os.MkdirTemp("", "cell-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Initialize IPFS client using embedded IPFSEnv
	if err = env.IPFSEnv.Boot(addr); err != nil {
		return err
	}

	// Initialize libp2p host
	env.Host, err = HostConfig{IPFS: env.IPFS}.New()
	return err
}

func (env *Env) Close() error {
	var errors []error

	// Close libp2p host independently
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

	// Check if the imported path is a directory
	if isDirectory(importedPath) {
		// If it's a directory, look for an executable file
		executablePath, err := env.findExecutableInDirectory(importedPath)
		if err != nil {
			return "", fmt.Errorf("IPFS path points to a directory but no executable found: %w", err)
		}
		return executablePath, nil
	}

	return importedPath, nil
}

// findExecutableInDirectory searches for an executable file in a directory
func (env *Env) findExecutableInDirectory(dirPath string) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	// Look for common executable names first
	commonExecutables := []string{"main", "app", "bin", "exec", "run", "start"}
	for _, name := range commonExecutables {
		path := filepath.Join(dirPath, name)
		if isExecutable(path) {
			return path, nil
		}
	}

	// If no common names found, look for any executable file
	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(dirPath, entry.Name())
			if isExecutable(path) {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("no executable file found in directory %s", dirPath)
}

// isDirectory checks if a path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// isExecutable checks if a file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
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

type HostConfig struct {
	IPFS    iface.CoreAPI
	Options []libp2p.Option
}

func (cfg HostConfig) New() (host.Host, error) {
	return libp2p.New(cfg.CombinedOptions()...)
}

func (cfg HostConfig) CombinedOptions() []libp2p.Option {
	return append(cfg.DefaultOptions(), cfg.Options...)
}

func (c HostConfig) DefaultOptions() []libp2p.Option {
	return []libp2p.Option{
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/2020"),
	}
}
