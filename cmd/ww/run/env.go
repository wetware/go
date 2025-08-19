package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/wetware/go/util"
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
	IPFS iface.CoreAPI
	Host host.Host
}

func (env *Env) Boot(addr string) (err error) {
	for _, bind := range []func() error{
		func() (err error) {
			env.IPFS, err = util.LoadIPFSFromName(addr)
			return
		},
		func() (err error) {
			env.Host, err = HostConfig{IPFS: env.IPFS}.New()
			return
		},
	} {
		if err = bind(); err != nil {
			break
		}
	}

	return
}

func (env *Env) Close() error {
	if env.Host != nil {
		return env.Host.Close()
	}

	return nil
}

// ResolveExecPath resolves an executable path, handling both IPFS paths and local filesystem paths.
// For IPFS paths, it downloads the file to the specified directory and makes it executable.
// For local paths, it resolves relative paths to absolute paths.
func (env *Env) ResolveExecPath(ctx context.Context, dir string, name string) (string, error) {
	if p, err := path.NewPath(name); err == nil {
		// Get the file from IPFS
		node, err := env.IPFS.Unixfs().Get(ctx, p)
		if err != nil {
			return "", fmt.Errorf("failed to get IPFS path: %w", err)
		}

		// Create target file path and update the 'name' variable.
		name = filepath.Join(dir, filepath.Base(p.String()))

		// Write the file to disk
		if err := files.WriteTo(node, name); err != nil {
			return "", fmt.Errorf("failed to write file: %w", err)
		}

		// Make executable
		if err := os.Chmod(name, 0755); err != nil {
			return "", fmt.Errorf("failed to make file executable: %w", err)
		}
	} else {
		// Handle non-IPFS paths - resolve relative paths to absolute
		if !filepath.IsAbs(name) {
			absPath, err := filepath.Abs(name)
			if err != nil {
				return "", fmt.Errorf("failed to resolve path %s: %w", name, err)
			}
			name = absPath
		}
	}

	return name, nil
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
