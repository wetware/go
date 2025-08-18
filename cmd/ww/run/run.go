package run

import (
	"context"
	"fmt"
	"os"
	os_exec "os/exec"
	"path/filepath"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/libp2p/go-libp2p/core/network"
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
			&cli.StringSliceFlag{
				Name:    "env",
				EnvVars: []string{"WW_ENV"},
			},
		}, flags.CapabilityFlags()...),

		// Environment hooks.
		////
		Before: func(c *cli.Context) error {
			return env.Boot(c.String("ipfs"))
		},
		After: func(c *cli.Context) error {
			return env.Close()
		},

		// Main
		////
		Action: func(c *cli.Context) error {
			dir, err := os.MkdirTemp("", "cell-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(dir)

			return Main(c, dir)
		},
	}
}

func Main(c *cli.Context, dir string) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	// Set up the RPC socket for the cell
	////
	host, guest, err := system.SocketConfig{
		Membrane: &system.Membrane{},
	}.New(ctx)
	if err != nil {
		return err
	}
	defer host.Close()
	defer guest.Close()

	// Check if first arg is an IPFS path and prepare name for CommandContext
	name, err := resolveExecPath(ctx, dir, c.Args().First())
	if err != nil {
		return err
	}

	// Run target in jailed subprocess
	////
	cmd := os_exec.CommandContext(ctx, name, c.Args().Tail()...)
	cmd.Dir = dir
	cmd.Env = c.StringSlice("env")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = sysProcAttr(dir)
	cmd.ExtraFiles = []*os.File{guest}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Cancel()

	// Set up libp2p protocol handler
	////
	env.Host.SetStreamHandler("/ww/0.1.0", func(s network.Stream) {
		defer s.Close()

		conn := rpc.NewConn(rpc.NewPackedStreamTransport(s), &rpc.Options{
			BaseContext:     func() context.Context { return ctx },
			BootstrapClient: host.Bootstrap(ctx), // cell-provided capability
		})
		defer conn.Close()

		select {
		case <-ctx.Done():
		case <-conn.Done():
		}
	})
	defer env.Host.RemoveStreamHandler("/ww/0.1.0")

	return cmd.Wait()
}

func resolveExecPath(ctx context.Context, dir string, name string) (string, error) {
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
	}

	return name, nil
}
