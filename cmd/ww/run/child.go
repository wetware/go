package run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ipfs/boxo/path"
	"github.com/wetware/go/system"
)

// ChildSourceType represents the type of source for a child cell
type ChildSourceType int

const (
	IPFSSource ChildSourceType = iota
	LocalSource
	FDSource
)

// ChildSpec represents a child cell specification
type ChildSpec struct {
	Name   string
	Source interface{} // IPFS path, local path, or FD number
	Type   ChildSourceType
}

// ChildManager handles child cell operations
type ChildManager struct {
	children  []*ChildSpec
	processes []*exec.Cmd
}

// NewChildManager creates a new child manager
func NewChildManager() *ChildManager {
	return &ChildManager{
		children:  make([]*ChildSpec, 0),
		processes: make([]*exec.Cmd, 0),
	}
}

// ParseWithChildFlag parses the --with-child flag value
func (cm *ChildManager) ParseWithChildFlag(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid --with-child format: must be name=source")
	}

	name := parts[0]
	source := parts[1]

	if name == "" {
		return fmt.Errorf("child name cannot be empty")
	}

	// Check for duplicate names
	for _, child := range cm.children {
		if child.Name == name {
			return fmt.Errorf("duplicate child name: %s", name)
		}
	}

	// Try to parse as IPFS path first
	if ipfsPath, err := path.NewPath(source); err == nil {
		cm.children = append(cm.children, &ChildSpec{
			Name:   name,
			Source: ipfsPath,
			Type:   IPFSSource,
		})
		return nil
	}

	// Try to parse as FD number
	if fd, err := strconv.Atoi(source); err == nil {
		// Validate FD is not stdio (0, 1, 2)
		if fd <= 2 {
			return fmt.Errorf("cannot use stdio file descriptors (0, 1, 2): %d", fd)
		}
		cm.children = append(cm.children, &ChildSpec{
			Name:   name,
			Source: fd,
			Type:   FDSource,
		})
		return nil
	}

	// Treat as local filesystem path
	cm.children = append(cm.children, &ChildSpec{
		Name:   name,
		Source: source,
		Type:   LocalSource,
	})
	return nil
}

// StartChildren starts all child cells
func (cm *ChildManager) StartChildren(ctx context.Context, env *Env) error {
	for i, child := range cm.children {
		if err := cm.startChild(ctx, child, env, i); err != nil {
			return fmt.Errorf("failed to start child %s: %w", child.Name, err)
		}
	}
	return nil
}

// startChild starts a single child cell
func (cm *ChildManager) startChild(ctx context.Context, child *ChildSpec, env *Env, index int) error {
	var execPath string
	var err error

	switch child.Type {
	case IPFSSource:
		// Download from IPFS
		ipfsPath := child.Source.(path.Path)
		execPath, err = env.ResolveExecPath(ctx, ipfsPath.String())
		if err != nil {
			return fmt.Errorf("failed to resolve IPFS path: %w", err)
		}

	case LocalSource:
		// Use local filesystem path
		localPath := child.Source.(string)
		absPath, err := filepath.Abs(localPath)
		if err != nil {
			return fmt.Errorf("failed to resolve local path: %w", err)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("local path does not exist: %s", absPath)
		}
		execPath = absPath

	case FDSource:
		// Use existing file descriptor
		fd := child.Source.(int)
		file := os.NewFile(uintptr(fd), fmt.Sprintf("child-%s", child.Name))
		if file == nil {
			return fmt.Errorf("invalid file descriptor: %d", fd)
		}
		// For FD sources, we'll handle this differently
		// The FD will be passed through to the child
		execPath = fmt.Sprintf("fd:%d", fd)
	}

	// Set up the RPC socket for the child cell
	host, guest, err := system.SocketConfig{
		Membrane: &system.Membrane{},
	}.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create socket pair for child %s: %w", child.Name, err)
	}
	defer host.Close()

	// Create command for child
	cmd := exec.CommandContext(ctx, execPath)
	cmd.Dir = env.Dir

	// Set up environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("WW_CHILD_NAME=%s", child.Name),
		fmt.Sprintf("WW_CHILD_INDEX=%d", index),
	)

	// Set up stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set up sysproc attributes for jailed execution
	cmd.SysProcAttr = sysProcAttr(env.Dir)

	// Add guest socket to extra files
	cmd.ExtraFiles = []*os.File{guest}

	// Start the child process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start child %s: %w", child.Name, err)
	}

	// Store the command for cleanup
	cm.processes = append(cm.processes, cmd)

	// Store the host socket for capability passing
	// TODO: Implement capability passing between parent and child

	return nil
}

// StopChildren stops all child cells
func (cm *ChildManager) StopChildren() {
	for _, cmd := range cm.processes {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
}

// GetChildren returns the list of child specifications
func (cm *ChildManager) GetChildren() []*ChildSpec {
	return cm.children
}
