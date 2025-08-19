package importcmd

import (
	"fmt"
	"path/filepath"

	"github.com/ipfs/boxo/path"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/cmd/internal/flags"
	"github.com/wetware/go/util"
)

var env util.IPFSEnv

func Command() *cli.Command {
	return &cli.Command{
		Name:      "import",
		ArgsUsage: "<ipfs-path> [local-path]",
		Usage:     "Import files and directories from IPFS to local filesystem",
		Description: `Import files and directories from IPFS to the local filesystem.
		
The command will:
1. Download the content from the specified IPFS path
2. Save it to the local filesystem (default: current directory)
3. Handle both files and directories recursively

Examples:
  ww import /ipfs/QmHash.../file.txt
  ww import /ipfs/QmHash.../directory ./local-dir
  ww import /ipfs/QmHash... .  # Import to current directory`,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Value:   "/dns4/localhost/tcp/5001/http",
				Usage:   "IPFS API endpoint",
			},
			&cli.BoolFlag{
				Name:    "executable",
				Aliases: []string{"x"},
				Usage:   "Make imported files executable (chmod +x)",
			},
		}, flags.CapabilityFlags()...),

		Before: func(c *cli.Context) error {
			return env.Boot(c.String("ipfs"))
		},
		After: func(c *cli.Context) error {
			return env.Close()
		},

		Action: Main,
	}
}

func Main(c *cli.Context) error {
	ctx := c.Context

	if c.NArg() < 1 || c.NArg() > 2 {
		return cli.Exit("import requires one or two arguments: <ipfs-path> [local-path]", 1)
	}

	ipfsPath := c.Args().Get(0)
	if ipfsPath == "" {
		return cli.Exit("IPFS path cannot be empty", 1)
	}

	// Parse IPFS path
	p, err := path.NewPath(ipfsPath)
	if err != nil {
		return cli.Exit(fmt.Sprintf("invalid IPFS path: %s", err), 1)
	}

	// Determine local destination path
	localPath := "."
	if c.NArg() == 2 {
		localPath = c.Args().Get(1)
	}

	// Resolve local path to absolute
	absLocalPath, err := filepath.Abs(localPath)
	if err != nil {
		return fmt.Errorf("failed to resolve local path %s: %w", localPath, err)
	}

	// Import from IPFS
	err = env.ImportFromIPFS(ctx, p, absLocalPath, c.Bool("executable"))
	if err != nil {
		return fmt.Errorf("failed to import from IPFS: %w", err)
	}

	fmt.Printf("Successfully imported %s to %s\n", ipfsPath, absLocalPath)
	return nil
}
