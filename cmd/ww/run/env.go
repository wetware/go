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

// ResolveIPFSPath resolves an IPFS path using the import functionality
func (env *Env) ResolveIPFSPath(ctx context.Context, ipfsPath path.Path) (string, error) {
	// Get IPFS client
	ipfs, err := env.GetIPFS()
	if err != nil {
		return "", err
	}

	// Get the node from IPFS
	node, err := ipfs.Unixfs().Get(ctx, ipfsPath)
	if err != nil {
		return "", fmt.Errorf("failed to get IPFS path: %w", err)
	}

	// Handle different node types
	switch node := node.(type) {
	case files.Directory:
		return env.ResolveIPFSDirectory(ctx, node, ipfsPath.String())
	case files.Node:
		return env.ResolveIPFSFile(ctx, node, ipfsPath.String())
	default:
		return "", fmt.Errorf("unexpected node type: %T", node)
	}
}

// ResolveIPFSFile handles IPFS file nodes, downloading them and making them executable
func (env *Env) ResolveIPFSFile(ctx context.Context, node files.Node, ipfsPath string) (string, error) {
	// Create target file path
	targetPath := filepath.Join(env.Dir, filepath.Base(ipfsPath))

	// Write the file to disk
	if err := files.WriteTo(node, targetPath); err != nil {
		return "", fmt.Errorf("failed to write IPFS file: %w", err)
	}

	// Make executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make file executable: %w", err)
	}

	return targetPath, nil
}

// ResolveIPFSDirectory handles IPFS directory nodes, looking for bin/ subdirectory with Go OS/arch convention
func (env *Env) ResolveIPFSDirectory(ctx context.Context, node files.Node, ipfsPath string) (string, error) {
	// Get OS and architecture from environment or runtime
	osName := env.OS()
	archName := env.Arch()

	// Look for bin/ subdirectory
	binDir := filepath.Join(env.Dir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Try to find executable in bin/OS_ARCH/ subdirectory
	osArchDir := filepath.Join(binDir, fmt.Sprintf("%s_%s", osName, archName))
	if err := os.MkdirAll(osArchDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create OS/arch directory: %w", err)
	}

	// Extract the bin/ directory from IPFS
	if err := env.extractIPFSDirectory(ctx, node, binDir); err != nil {
		return "", fmt.Errorf("failed to extract IPFS directory: %w", err)
	}

	// Look for executable in the OS/arch subdirectory
	executablePath := filepath.Join(osArchDir, filepath.Base(ipfsPath))
	if _, err := os.Stat(executablePath); err == nil {
		// Make executable
		if err := os.Chmod(executablePath, 0755); err != nil {
			return "", fmt.Errorf("failed to make file executable: %w", err)
		}
		return executablePath, nil
	}

	// Fallback: look for executable directly in bin/ directory
	fallbackPath := filepath.Join(binDir, filepath.Base(ipfsPath))
	if _, err := os.Stat(fallbackPath); err == nil {
		// Make executable
		if err := os.Chmod(fallbackPath, 0755); err != nil {
			return "", fmt.Errorf("failed to make file executable: %w", err)
		}
		return fallbackPath, nil
	}

	return "", fmt.Errorf("no executable found in IPFS directory %s", ipfsPath)
}

// extractIPFSDirectory recursively extracts an IPFS directory to the local filesystem
func (env *Env) extractIPFSDirectory(ctx context.Context, node files.Node, targetDir string) error {
	iter := node.(files.DirIterator)
	for iter.Next() {
		child := iter.Node()
		childName := iter.Name()
		childPath := filepath.Join(targetDir, childName)

		if _, ok := child.(files.Directory); ok {
			// Create subdirectory and recurse
			if err := os.MkdirAll(childPath, 0755); err != nil {
				return fmt.Errorf("failed to create subdirectory %s: %w", childPath, err)
			}
			if err := env.extractIPFSDirectory(ctx, child, childPath); err != nil {
				return err
			}
		} else {
			// Extract file
			if err := files.WriteTo(child, childPath); err != nil {
				return fmt.Errorf("failed to write file %s: %w", childPath, err)
			}
		}
	}
	return nil
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
