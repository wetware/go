package export

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ipfs/boxo/files"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/cmd/internal/flags"
	"github.com/wetware/go/util"
)

var env Env

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

type Env struct {
	IPFS iface.CoreAPI
}

func (env *Env) Boot(addr string) error {
	var err error
	env.IPFS, err = util.LoadIPFSFromName(addr)
	return err
}

func (env *Env) Close() error {
	// No cleanup needed for IPFS client
	return nil
}

// AddToIPFS adds a file or directory to IPFS recursively
func (env *Env) AddToIPFS(ctx context.Context, localPath string) (string, error) {
	// Get file info to determine if it's a directory
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat %s: %w", localPath, err)
	}

	var node files.Node
	if fileInfo.IsDir() {
		// Handle directory
		node, err = env.CreateDirectoryNode(ctx, localPath)
		if err != nil {
			return "", fmt.Errorf("failed to create directory node: %w", err)
		}
	} else {
		// Handle single file
		node, err = env.CreateFileNode(ctx, localPath)
		if err != nil {
			return "", fmt.Errorf("failed to create file node: %w", err)
		}
	}

	// Check if IPFS client is available
	if env.IPFS == nil {
		return "", fmt.Errorf("IPFS client not initialized")
	}

	// Add the node to IPFS using Unixfs API
	path, err := env.IPFS.Unixfs().Add(ctx, node)
	if err != nil {
		return "", fmt.Errorf("failed to add to IPFS: %w", err)
	}

	return path.String(), nil
}

// CreateFileNode creates a files.Node for a single file
func (env *Env) CreateFileNode(ctx context.Context, filePath string) (files.Node, error) {
	// Read the file content into memory
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Create a file node from the content
	return files.NewBytesFile(content), nil
}

// CreateDirectoryNode creates a files.Node for a directory recursively
func (env *Env) CreateDirectoryNode(ctx context.Context, dirPath string) (files.Node, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	// Create a map to hold directory contents
	dirMap := make(map[string]files.Node)

	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// Recursively handle subdirectories
			childNode, err := env.CreateDirectoryNode(ctx, entryPath)
			if err != nil {
				return nil, err
			}

			// Add subdirectory to the map
			dirMap[entry.Name()] = childNode
		} else {
			// Handle files
			childNode, err := env.CreateFileNode(ctx, entryPath)
			if err != nil {
				return nil, err
			}

			// Add file to the map
			dirMap[entry.Name()] = childNode
		}
	}

	// Create directory from the map
	return files.NewMapDirectory(dirMap), nil
}
