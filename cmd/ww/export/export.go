package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/wetware/go/cmd/internal/flags"
	"github.com/wetware/go/util"
)

var env util.IPFSEnv

func Command() *cli.Command {
	return &cli.Command{
		Name:      "export",
		ArgsUsage: "<path>",
		Usage:     "Add a file or directory to IPFS recursively",
		Description: `Add a file or directory to IPFS recursively, equivalent to 'ipfs add -r <path>'.
		
The command will:
1. Read the specified path from the local filesystem
2. Add it to IPFS recursively (including all subdirectories and files)
3. Print the resulting IPFS path to stdout followed by a newline

Examples:
  ww export /path/to/file.txt
  ww export /path/to/directory
  ww export .  # Export current directory`,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Value:   "/dns4/localhost/tcp/5001/http",
				Usage:   "IPFS API endpoint",
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

	if c.NArg() != 1 {
		return cli.Exit("export requires exactly one argument: <path>", 1)
	}

	argPath := c.Args().First()
	if argPath == "" {
		return cli.Exit("path cannot be empty", 1)
	}

	// Resolve the path to absolute
	absPath, err := filepath.Abs(argPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path %s: %w", argPath, err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return cli.Exit(fmt.Sprintf("path does not exist: %s", absPath), 1)
	}

	// Add the file/directory to IPFS
	ipfsPath, err := env.AddToIPFS(ctx, absPath)
	if err != nil {
		return fmt.Errorf("failed to add to IPFS: %w", err)
	}

	// Print the IPFS path to stdout followed by a newline
	fmt.Println(ipfsPath)
	return nil
}
