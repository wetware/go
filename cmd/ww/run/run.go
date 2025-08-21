package run

import (
	"context"
	"fmt"
	"os"
	os_exec "os/exec"

	"capnproto.org/go/capnp/v3/rpc"
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
			&cli.StringSliceFlag{
				Name:     "fd",
				Category: "FILE DESCRIPTORS",
				Usage:    "map existing parent fd to name (e.g., db=3). Use --fd multiple times for multiple fds.",
				// Note: Each --fd flag should be name=fdnum
				// Use --fd-map for target fd, --fd-verbose for logging
			},
			&cli.StringSliceFlag{
				Name:     "fd-map",
				Category: "FILE DESCRIPTORS",
				Usage:    "override numeric target for a named fd (e.g., db=10)",
			},
			&cli.StringSliceFlag{
				Name:     "fdctl",
				Category: "FILE DESCRIPTORS",
				Usage:    "accept fds via SCM_RIGHTS or inherit existing (e.g., /path/to/sock or inherit:5)",
			},
			&cli.StringFlag{
				Name:     "use-systemd-fds",
				Category: "FILE DESCRIPTORS",
				Usage:    "import fds from systemd socket activation (optional prefix)",
			},
			&cli.StringFlag{
				Name:     "fd-from",
				Category: "FILE DESCRIPTORS",
				Usage:    "bulk capability imports from file or stdin (e.g., @imports or -)",
			},
			&cli.BoolFlag{
				Name:     "fd-verbose",
				Category: "FILE DESCRIPTORS",
				Usage:    "verbose logging of fd grants",
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
		Action: Main,
	}
}

func Main(c *cli.Context) error {
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

	// Initialize fd manager
	fdManager := NewFDManager(c.Bool("fd-verbose"))

	// Process fd-related flags
	if err := processFDFlags(c, fdManager); err != nil {
		return fmt.Errorf("fd flag processing failed: %w", err)
	}

	// Check if first arg is an IPFS path and prepare name for CommandContext
	name, err := env.ResolveExecPath(ctx, c.Args().First())
	if err != nil {
		return err
	}

	// Prepare file descriptors for child process
	fdFiles, err := fdManager.PrepareFDs()
	if err != nil {
		return fmt.Errorf("failed to prepare file descriptors: %w", err)
	}
	defer fdManager.Close()

	// Create symlinks in jail if requested
	if err := fdManager.CreateSymlinks(env.Dir); err != nil {
		return fmt.Errorf("failed to create symlinks: %w", err)
	}

	// Run target in jailed subprocess
	////
	cmd := os_exec.CommandContext(ctx, name, c.Args().Tail()...)
	cmd.Dir = env.Dir

	// Combine environment variables
	baseEnv := c.StringSlice("env")
	fdEnv := fdManager.GenerateEnvVars()
	cmd.Env = append(baseEnv, fdEnv...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = sysProcAttr(env.Dir)

	// Combine extra files: guest socket + fd files
	extraFiles := []*os.File{guest}
	extraFiles = append(extraFiles, fdFiles...)
	cmd.ExtraFiles = extraFiles

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

// processFDFlags processes all fd-related command line flags
func processFDFlags(c *cli.Context, fdManager *FDManager) error {
	// Process --fd flags (can be specified multiple times)
	for _, fdFlag := range c.StringSlice("fd") {
		if err := fdManager.ParseFDFlag(fdFlag); err != nil {
			return fmt.Errorf("invalid --fd flag '%s': %w", fdFlag, err)
		}
	}

	// Process --fd-map flags
	for _, fdMapFlag := range c.StringSlice("fd-map") {
		if err := fdManager.ParseFDMapFlag(fdMapFlag); err != nil {
			return fmt.Errorf("invalid --fd-map flag '%s': %w", fdMapFlag, err)
		}
	}

	// Process --fdctl flags
	for _, fdctlFlag := range c.StringSlice("fdctl") {
		if err := fdManager.ParseFDCTLFlag(fdctlFlag); err != nil {
			return fmt.Errorf("invalid --fdctl flag '%s': %w", fdctlFlag, err)
		}
	}

	// Process --use-systemd-fds flag
	if systemdPrefix := c.String("use-systemd-fds"); systemdPrefix != "" {
		if err := fdManager.UseSystemdFDs(systemdPrefix); err != nil {
			return fmt.Errorf("systemd fd import failed: %w", err)
		}
	}

	// Process --fd-from flag
	if fdFromPath := c.String("fd-from"); fdFromPath != "" {
		if err := fdManager.ParseFDFromFile(fdFromPath); err != nil {
			return fmt.Errorf("fd spec file parsing failed: %w", err)
		}
	}

	return nil
}
