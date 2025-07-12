package pull

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "pull",
		Usage:     "fetch a package from IPFS",
		ArgsUsage: "<image.tar>",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "path",
				EnvVars: []string{"WW_PATH"},
				Value:   "~/.ww",
				Usage:   "path to local wetware storage",
			},
		},
		Before: setup,
		Action: Main,
		// After:     teardown,
	}
}

func Main(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return cli.Exit("missing IPFS hash argument", 1)
	}

	// Get and expand the storage path
	wwpath := c.String("path")
	if strings.HasPrefix(wwpath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		wwpath = filepath.Join(homeDir, wwpath[2:])
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(wwpath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory %s: %w", wwpath, err)
	}

	for _, hash := range c.Args().Slice() {
		// Use ipfs get to download the content to the storage directory
		cmd := exec.Command("ipfs", "get", hash)
		cmd.Dir = wwpath // Set working directory to storage path
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("exec: failed to get %s: %v\nOutput: %s", hash, err, string(output))
		}

		fmt.Printf("Successfully pulled %s to %s\n", hash, wwpath)
	}

	return nil
}
