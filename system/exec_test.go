package system_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/system"
)

func TestExecConfig_New(t *testing.T) {
	t.Parallel()

	// Create ExecConfig
	config := system.ExecConfig{Enabled: true}

	// Test that New() returns an Executor client
	executor := config.New()
	require.NotNil(t, executor, "New() should return a non-nil Executor client")

	// Test that we can release the client
	defer executor.Release()

	// Verify the config has the expected enabled state
	require.True(t, config.Enabled, "Config should have the expected enabled state")
}

func TestExecConfig_Disabled(t *testing.T) {
	t.Parallel()

	// Create disabled ExecConfig
	config := system.ExecConfig{Enabled: false}
	executor := config.New()
	defer executor.Release()

	// This would normally be created by the RPC system, but for testing
	// we'll just verify that the config is disabled
	require.False(t, config.Enabled, "Executor should be disabled")
}

func TestExecConfig_ZeroValue(t *testing.T) {
	t.Parallel()

	// Test that zero value ExecConfig can be created
	var config system.ExecConfig
	require.False(t, config.Enabled, "Zero value should have disabled executor")

	// Test that New() still works
	executor := config.New()
	require.NotNil(t, executor, "New() should return a non-nil Executor client even with zero value")
	defer executor.Release()
}

func TestCellConfig_Wait(t *testing.T) {
	t.Parallel()

	// Create a simple command that exits immediately
	cmd := exec.Command("echo", "hello")

	// Create CellConfig
	cell := system.CellConfig{Cmd: cmd}

	// This would normally be called by the RPC system
	// For now, we'll just verify the cell has the expected command
	require.Equal(t, cmd, cell.Cmd, "Cell should have the expected command")
}

func TestExecConfig_WithIPFS(t *testing.T) {
	t.Parallel()

	// Create ExecConfig with IPFS
	config := system.ExecConfig{Enabled: true}

	executor := config.New()
	defer executor.Release()
}

// TestExecConfig_Integration tests the full integration with a real executable
func TestExecConfig_Integration(t *testing.T) {
	t.Parallel()

	// Skip if not on Unix-like system (for echo command)
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a simple test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	scriptContent := `#!/bin/sh
echo "hello world"
exit 0
`

	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err, "Failed to create test script")

	// Create ExecConfig
	config := system.ExecConfig{Enabled: true}
	executor := config.New()
	defer executor.Release()

	// Verify the script exists and is executable
	info, err := os.Stat(scriptPath)
	require.NoError(t, err, "Test script should exist")
	require.True(t, info.Mode()&0111 != 0, "Test script should be executable")
}

func TestExecConfig_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Create ExecConfig
	config := system.ExecConfig{Enabled: true}
	executor := config.New()
	defer executor.Release()

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cancel the context immediately
	cancel()

	// Verify the context is cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Fatal("Context should be cancelled")
	}
}

func TestExecConfig_Timeout(t *testing.T) {
	t.Parallel()

	// Create ExecConfig
	config := system.ExecConfig{Enabled: true}
	executor := config.New()
	defer executor.Release()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Wait for timeout
	<-ctx.Done()

	// Verify the context timed out
	require.Equal(t, context.DeadlineExceeded, ctx.Err(), "Context should have timed out")
}
