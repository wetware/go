package importcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/cmd/internal/flags"
	"github.com/wetware/go/util"
)

var env Env

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

// ImportFromIPFS imports content from IPFS to the local filesystem
func (env *Env) ImportFromIPFS(ctx context.Context, ipfsPath path.Path, localPath string, makeExecutable bool) error {
	// Check if IPFS client is available
	if env.IPFS == nil {
		return fmt.Errorf("IPFS client not initialized")
	}

	// Get the node from IPFS
	node, err := env.IPFS.Unixfs().Get(ctx, ipfsPath)
	if err != nil {
		return fmt.Errorf("failed to get IPFS path: %w", err)
	}

	// Handle different node types
	switch node := node.(type) {
	case files.Directory:
		return env.importIPFSDirectory(ctx, node, ipfsPath.String(), localPath, makeExecutable)
	case files.Node:
		return env.importIPFSFile(ctx, node, ipfsPath.String(), localPath, makeExecutable)
	default:
		return fmt.Errorf("unexpected node type: %T", node)
	}
}

// importIPFSFile handles importing a single file from IPFS
func (env *Env) importIPFSFile(ctx context.Context, node files.Node, ipfsPath, localPath string, makeExecutable bool) error {
	// Determine target file path
	var targetPath string
	if isDirectory(localPath) {
		// If localPath is a directory, use the filename from IPFS path
		targetPath = filepath.Join(localPath, filepath.Base(ipfsPath))
	} else {
		// If localPath is a file path, use it directly
		targetPath = localPath
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Write the file to disk
	if err := files.WriteTo(node, targetPath); err != nil {
		return fmt.Errorf("failed to write IPFS file: %w", err)
	}

	// Make executable if requested
	if makeExecutable {
		if err := os.Chmod(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to make file executable: %w", err)
		}
	}

	return nil
}

// importIPFSDirectory handles importing a directory from IPFS
func (env *Env) importIPFSDirectory(ctx context.Context, node files.Node, ipfsPath, localPath string, makeExecutable bool) error {
	// Ensure local path is a directory
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Extract the directory recursively
	return env.extractIPFSDirectory(ctx, node, localPath, makeExecutable)
}

// extractIPFSDirectory recursively extracts an IPFS directory to the local filesystem
func (env *Env) extractIPFSDirectory(ctx context.Context, node files.Node, targetDir string, makeExecutable bool) error {
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
			if err := env.extractIPFSDirectory(ctx, child, childPath, makeExecutable); err != nil {
				return err
			}
		} else {
			// Extract file
			if err := files.WriteTo(child, childPath); err != nil {
				return fmt.Errorf("failed to write file %s: %w", childPath, err)
			}

			// Make executable if requested
			if makeExecutable {
				if err := os.Chmod(childPath, 0755); err != nil {
					return fmt.Errorf("failed to make file executable: %w", err)
				}
				return nil
			}
		}
	}
	return nil
}

// isDirectory checks if the given path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
