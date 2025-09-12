package lang

import (
	"fmt"
	"io"
	"strings"

	"github.com/ipfs/boxo/files"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

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
	case ":type":
		return builtin.String(n.Type()), nil
	case ":size":
		size, err := n.Size()
		if err != nil {
			return nil, fmt.Errorf("failed to get size: %w", err)
		}
		return builtin.Int64(size), nil
	case ":is-file":
		return builtin.Bool(n.Type() == "file"), nil
	case ":is-directory":
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
	case ":read":
		// Read all content
		content, err := io.ReadAll(f.File)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		return content, nil
	case ":read-string":
		// Read content as string
		content, err := io.ReadAll(f.File)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		return builtin.String(string(content)), nil
	case ":size":
		size, err := f.Size()
		if err != nil {
			return nil, fmt.Errorf("failed to get size: %w", err)
		}
		return builtin.Int64(size), nil
	case ":type":
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
	case ":list":
		// List directory entries
		entries := make([]core.Any, 0)
		it := d.Entries()
		for it.Next() {
			entries = append(entries, builtin.String(it.Name()))
		}
		return builtin.NewList(entries...), nil
	case ":entries":
		// Return directory entries as a list of strings
		entries := make([]core.Any, 0)
		it := d.Entries()
		for it.Next() {
			entries = append(entries, builtin.String(it.Name()))
		}
		return builtin.NewList(entries...), nil
	case ":size":
		size, err := d.Size()
		if err != nil {
			return nil, fmt.Errorf("failed to get size: %w", err)
		}
		return builtin.Int64(size), nil
	case ":type":
		return builtin.String("directory"), nil
	default:
		return nil, fmt.Errorf("unknown directory method: %s", method)
	}
}
