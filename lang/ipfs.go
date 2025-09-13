package lang

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

var _ core.Invokable = (*IPFS)(nil)

type IPFS struct {
	iface.CoreAPI
}

// IPFS methods: (ipfs :cat /ipfs/Qm...) or (ipfs :get /ipfs/Qm...)
func (i IPFS) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("ipfs requires at least 2 arguments: (ipfs :method path)")
	}

	// First argument should be a keyword (:cat or :get)
	method, ok := args[0].(builtin.Keyword)
	if !ok {
		return nil, fmt.Errorf("ipfs method must be a keyword, got %T", args[0])
	}

	src, ok := args[1].(path.Path)
	if !ok {
		return nil, fmt.Errorf("ipfs path must be a Path object, got %T", args[1])
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	switch method {
	case "cat":
		return i.Cat(ctx, src)
	case "get":
		return i.Get(ctx, src)
	default:
		return nil, fmt.Errorf("unknown ipfs method: %s (supported: :cat, :get)", method)
	}
}

// Cat returns the content of an IPFS file as []byte
func (i IPFS) Cat(ctx context.Context, p path.Path) (core.Any, error) {
	if i.CoreAPI == nil || i.CoreAPI.Unixfs() == nil {
		return nil, fmt.Errorf("IPFS client not initialized")
	}

	// Get the node from IPFS
	node, err := i.Unixfs().Get(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPFS path: %w", err)
	}

	// Handle different node types
	switch node := node.(type) {
	case files.File:
		// For files, read the content into a byte slice
		content, err := io.ReadAll(node)
		if err != nil {
			return nil, fmt.Errorf("failed to read file content: %w", err)
		}
		return content, nil
	case files.Directory:
		return nil, fmt.Errorf("path is a directory, use :get instead of :cat")
	default:
		return nil, fmt.Errorf("unexpected node type: %T", node)
	}
}

// Get returns the IPFS node as a file-like object (io.Reader) that can be used with eval
func (i IPFS) Get(ctx context.Context, p path.Path) (core.Any, error) {
	if i.CoreAPI == nil || i.CoreAPI.Unixfs() == nil {
		return nil, fmt.Errorf("IPFS client not initialized")
	}

	// Get the node from IPFS
	node, err := i.Unixfs().Get(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPFS path: %w", err)
	}

	switch node := node.(type) {
	case files.File:
		return File{File: node}, nil
	case files.Directory:
		return Directory{Directory: node}, nil
	default:
		return Node{Node: node}, nil
	}
}

type Node struct {
	files.Node
}

func (n Node) String() string {
	return fmt.Sprintf("<IPFS Node: %s>", n.Type())
}

func (n Node) Type() string {
	switch n.Node.(type) {
	case files.File:
		return "file"
	case files.Directory:
		return "directory"
	default:
		return "unknown"
	}
}

// Implement core.Invokable for Node
func (n Node) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		return n.String(), nil
	}

	method, ok := args[0].(builtin.Keyword)
	if !ok {
		return nil, fmt.Errorf("node method must be a keyword, got %T", args[0])
	}

	switch method {
	case "type":
		return builtin.String(n.Type()), nil
	case "size":
		size, err := n.Size()
		if err != nil {
			return nil, fmt.Errorf("failed to get size: %w", err)
		}
		return builtin.Int64(size), nil
	case "is-file":
		return builtin.Bool(n.Type() == "file"), nil
	case "is-directory":
		return builtin.Bool(n.Type() == "directory"), nil
	default:
		return nil, fmt.Errorf("unknown node method: %s", method)
	}
}

type File struct {
	files.File
}

func (f File) String() string {
	size, err := f.Size()
	if err != nil {
		return "<IPFS File: (size unknown)>"
	}
	return fmt.Sprintf("<IPFS File: %d bytes>", size)
}

// Implement core.Invokable for File
func (f File) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		return f.String(), nil
	}

	method, ok := args[0].(builtin.Keyword)
	if !ok {
		return nil, fmt.Errorf("file method must be a keyword, got %T", args[0])
	}

	switch method {
	case "read":
		// Read all content
		content, err := io.ReadAll(f.File)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		return content, nil
	case "read-string":
		// Read content as string
		content, err := io.ReadAll(f.File)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		return builtin.String(string(content)), nil
	case "size":
		size, err := f.Size()
		if err != nil {
			return nil, fmt.Errorf("failed to get size: %w", err)
		}
		return builtin.Int64(size), nil
	case "type":
		return builtin.String("file"), nil
	default:
		return nil, fmt.Errorf("unknown file method: %s", method)
	}
}

type Directory struct {
	files.Directory
}

func (d Directory) String() string {
	// Try to get directory entries for a more informative string
	entries := make([]string, 0)
	it := d.Entries()
	for it.Next() {
		entries = append(entries, it.Name())
	}

	if len(entries) == 0 {
		return "<IPFS Directory: (empty)>"
	}

	if len(entries) <= 3 {
		return fmt.Sprintf("<IPFS Directory: [%s]>", strings.Join(entries, ", "))
	}

	return fmt.Sprintf("<IPFS Directory: [%s, ... %d more]>", strings.Join(entries[:3], ", "), len(entries)-3)
}

// Implement core.Invokable for Directory
func (d Directory) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		return d.String(), nil
	}

	method, ok := args[0].(builtin.Keyword)
	if !ok {
		return nil, fmt.Errorf("directory method must be a keyword, got %T", args[0])
	}

	switch method {
	case "list":
		// List directory entries
		entries := make([]core.Any, 0)
		it := d.Entries()
		for it.Next() {
			entries = append(entries, builtin.String(it.Name()))
		}
		return builtin.NewList(entries...), nil
	case "entries":
		// Return directory entries as a list of strings
		entries := make([]core.Any, 0)
		it := d.Entries()
		for it.Next() {
			entries = append(entries, builtin.String(it.Name()))
		}
		return builtin.NewList(entries...), nil
	case "size":
		size, err := d.Size()
		if err != nil {
			return nil, fmt.Errorf("failed to get size: %w", err)
		}
		return builtin.Int64(size), nil
	case "type":
		return builtin.String("directory"), nil
	default:
		return nil, fmt.Errorf("unknown directory method: %s", method)
	}
}
