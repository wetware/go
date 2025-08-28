package run

import (
	"context"
	"fmt"
	"log/slog"
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

func (env *Env) Boot(addr string, port int) (err error) {
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
	env.Host, err = HostConfig{IPFS: env.IPFS, Port: port}.New()
	if err == nil {
		logBoot(env)
	}
	return err
}

func logBoot(env *Env) {
	args := make([]any, 0, len(env.Host.Addrs())+1)
	args = append(args, slog.String("id", env.Host.ID().String()))
	for _, addr := range env.Host.Addrs() {
		args = append(args, slog.String("addr", addr.String()))
	}
	slog.Info("LibP2P initialized", args...)
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

type HostConfig struct {
	IPFS    iface.CoreAPI
	Options []libp2p.Option
	Port    int
}

func (cfg HostConfig) New() (host.Host, error) {
	return libp2p.New(cfg.CombinedOptions()...)
}

func (cfg HostConfig) CombinedOptions() []libp2p.Option {
	return append(cfg.DefaultOptions(), cfg.Options...)
}

func (c HostConfig) DefaultOptions() []libp2p.Option {
	return []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", c.Port)),
	}
}
