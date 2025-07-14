package init

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/wetware/go/util"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:   "init",
		Action: configureEnv,
	}
}
func configureEnv(c *cli.Context) error {
	wwpath, err := util.ExpandHome(c.String("path"))
	if err != nil {
		return fmt.Errorf("failed to expand wetware path: %w", err)
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(wwpath, 0o755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create storage directory %s: %w", wwpath, err)
	}

	// Create mount points for /ipfs and /ipns under wwpath
	ipfsDir := filepath.Join(wwpath, "ipfs")
	ipnsDir := filepath.Join(wwpath, "ipns")
	if err := os.MkdirAll(ipfsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create ipfs mountpoint %s: %w", ipfsDir, err)
	}
	if err := os.MkdirAll(ipnsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create ipns mountpoint %s: %w", ipnsDir, err)
	}

	// Launch `ipfs mount` in the background
	cmd := exec.Command("ipfs",
		"mount",
		"--ipfs-path", ipfsDir,
		"--ipns-path", ipnsDir,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to mount IPFS FUSE: %w", err)
	}

	// Optionally: store cmd.Process to unmount later, or leave it running.
	return nil
}
