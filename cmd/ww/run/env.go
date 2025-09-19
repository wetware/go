package run

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
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
	NS   string
	MDNS bool
}

func (cfg EnvConfig) New() (env Env, err error) {
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
	env.Host, err = HostConfig{
		NS:   cfg.NS,
		IPFS: env.IPFS,
		Port: cfg.Port,
	}.New()
	if err != nil {
		err = fmt.Errorf("failed to create libp2p host: %w", err)
		return
	}

	// Initialize mDNS discovery service if enabled
	if cfg.MDNS {
		env.MDNS = mdns.NewMdnsService(env.Host, env.NS, &MDNSPeerHandler{
			Peerstore: env.Host.Peerstore(),
			TTL:       peerstore.AddressTTL,
		})
		env.MDNS.Start()
		slog.Info("mDNS discovery service started")
	}
	return
}

type Env struct {
	util.IPFSEnv
	Host host.Host
	NS   string
	Dir  string // Temporary directory for cell execution
	MDNS mdns.Service
}

func (env *Env) Close() error {
	var errors []error

	// Close mDNS service
	if env.MDNS != nil {
		if err := env.MDNS.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close mDNS service: %w", err))
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

type HostConfig struct {
	NS      string
	IPFS    iface.CoreAPI
	Options []libp2p.Option
	Port    int
}

type HostWithDiscovery struct {
	host.Host
	MDNS mdns.Service
}

// Start starts the mDNS discovery service
func (h *HostWithDiscovery) Start() {
	if h.MDNS != nil {
		h.MDNS.Start()
	}
}

// Close stops the mDNS discovery service and closes the host
func (h *HostWithDiscovery) Close() error {
	if h.MDNS != nil {
		h.MDNS.Close()
	}
	return h.Host.Close()
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

// mdnsNotifee implements the mdns.Notifee interface
type MDNSPeerHandler struct {
	peerstore.Peerstore
	TTL time.Duration
}

func (m MDNSPeerHandler) HandlePeerFound(pi peer.AddrInfo) {
	if m.TTL < 0 {
		m.TTL = peerstore.AddressTTL
	}
	slog.Info("mDNS discovered peer",
		"peer_id", pi.ID,
		"addrs", pi.Addrs,
		"ttl", m.TTL)
	m.Peerstore.AddAddrs(pi.ID, pi.Addrs, peerstore.AddressTTL)
}
