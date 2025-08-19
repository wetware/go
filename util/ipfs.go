package util

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	ma "github.com/multiformats/go-multiaddr"
)

func LoadIPFSFromName(name string) (iface.CoreAPI, error) {
	if name == "" {
		name = rpc.DefaultPathRoot
	}

	// attempt to load as multiaddr
	if a, err := ma.NewMultiaddr(name); err == nil {
		if api, err := rpc.NewApiWithClient(a, http.DefaultClient); err == nil {
			return api, nil
		}
	}

	// attempt to load as URL
	if u, err := url.ParseRequestURI(name); err == nil {
		return rpc.NewURLApiWithClient(u.String(), http.DefaultClient)
	}

	if api, err := rpc.NewPathApi(name); err == nil {
		return api, nil
	}

	return nil, fmt.Errorf("invalid ipfs addr: %s", name)
}

// IPFSEnv provides shared IPFS functionality for command environments
type IPFSEnv struct {
	IPFS iface.CoreAPI
}

// Boot initializes the IPFS client connection
func (env *IPFSEnv) Boot(addr string) error {
	var err error
	env.IPFS, err = LoadIPFSFromName(addr)
	return err
}

// Close cleans up the IPFS environment
func (env *IPFSEnv) Close() error {
	// No cleanup needed for IPFS client
	return nil
}

// GetIPFS returns the IPFS client, ensuring it's initialized
func (env *IPFSEnv) GetIPFS() (iface.CoreAPI, error) {
	if env.IPFS == nil {
		return nil, fmt.Errorf("IPFS client not initialized")
	}
	return env.IPFS, nil
}

// AddToIPFS adds a file or directory to IPFS recursively
func (env IPFSEnv) AddToIPFS(ctx context.Context, localPath string) (string, error) {
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

	// Get IPFS client
	ipfs, err := env.GetIPFS()
	if err != nil {
		return "", err
	}

	// Add the node to IPFS using Unixfs API
	path, err := ipfs.Unixfs().Add(ctx, node)
	if err != nil {
		return "", fmt.Errorf("failed to add to IPFS: %w", err)
	}

	return path.String(), nil
}

// CreateFileNode creates a files.Node for a single file
func (env IPFSEnv) CreateFileNode(ctx context.Context, filePath string) (files.Node, error) {
	// Read the file content into memory
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Create a file node from the content
	return files.NewBytesFile(content), nil
}

// CreateDirectoryNode creates a files.Node for a directory recursively
func (env IPFSEnv) CreateDirectoryNode(ctx context.Context, dirPath string) (files.Node, error) {
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

// ImportFromIPFS imports content from IPFS to the local filesystem
func (env *IPFSEnv) ImportFromIPFS(ctx context.Context, ipfsPath path.Path, localPath string, makeExecutable bool) error {
	// Get IPFS client
	ipfs, err := env.GetIPFS()
	if err != nil {
		return err
	}

	// Get the node from IPFS
	node, err := ipfs.Unixfs().Get(ctx, ipfsPath)
	if err != nil {
		return fmt.Errorf("failed to get IPFS path: %w", err)
	}

	// Handle different node types
	switch node := node.(type) {
	case files.Directory:
		return env.ImportIPFSDirectory(ctx, node, ipfsPath.String(), localPath, makeExecutable)
	case files.Node:
		return env.ImportIPFSFile(ctx, node, ipfsPath.String(), localPath, makeExecutable)
	default:
		return fmt.Errorf("unexpected node type: %T", node)
	}
}

// ImportIPFSFile handles importing a single file from IPFS
func (env *IPFSEnv) ImportIPFSFile(ctx context.Context, node files.Node, ipfsPath, localPath string, makeExecutable bool) error {
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

// ImportIPFSDirectory handles importing a directory from IPFS
func (env *IPFSEnv) ImportIPFSDirectory(ctx context.Context, node files.Node, ipfsPath, localPath string, makeExecutable bool) error {
	// Ensure local path is a directory
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Extract the directory recursively
	return env.ExtractIPFSDirectory(ctx, node, localPath, makeExecutable)
}

// ExtractIPFSDirectory recursively extracts an IPFS directory to the local filesystem
func (env *IPFSEnv) ExtractIPFSDirectory(ctx context.Context, node files.Node, targetDir string, makeExecutable bool) error {
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
			if err := env.ExtractIPFSDirectory(ctx, child, childPath, makeExecutable); err != nil {
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
			}
		}
	}
	return nil
}

// isDirectory checks if a path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
