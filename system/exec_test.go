package system_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/system"
)

func TestNewMembraneBasic(t *testing.T) {
	t.Parallel()

	// Test that we can create a membrane with basic command
	cmdArgs := []string{"echo", "hello", "world"}

	// We'll need to pass nil for IPFS since we can't easily mock the full interface
	// This will test the basic structure creation
	membrane, err := system.NewMembrane(system.IPFS{}, cmdArgs...)
	require.NoError(t, err)
	require.NotNil(t, membrane)
	require.NotNil(t, membrane.Cmd)
	require.NotNil(t, membrane.PrivateKey)
	require.NotEmpty(t, membrane.PeerID)
	// ipfsClient will be empty since we passed nil
	require.Equal(t, system.IPFS{}, membrane.IPFS)
}

func TestNewMembraneEmptyArgs(t *testing.T) {
	t.Parallel()

	cmdArgs := []string{}

	_, err := system.NewMembrane(system.IPFS{}, cmdArgs...)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no command specified")
}

func TestMembraneStartBasic(t *testing.T) {
	t.Parallel()

	cmdArgs := []string{"echo", "test"}

	membrane, err := system.NewMembrane(system.IPFS{}, cmdArgs...)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = membrane.Start(ctx)
	require.NoError(t, err)

	// Wait a bit for the process to start
	time.Sleep(100 * time.Millisecond)

	// Check that the process is running
	require.NotNil(t, membrane.Cmd.Process)
	require.True(t, membrane.Cmd.ProcessState == nil) // Process should still be running

	// Clean up
	membrane.Cmd.Process.Kill()
}

func TestMembraneStartInvalidCommand(t *testing.T) {
	t.Parallel()

	cmdArgs := []string{"nonexistent_command_that_should_not_exist_12345"}

	membrane, err := system.NewMembrane(system.IPFS{}, cmdArgs...)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = membrane.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to start cell")
}

func TestDefaultCellWait(t *testing.T) {
	t.Parallel()

	cmdArgs := []string{"echo", "test"}

	membrane, err := system.NewMembrane(system.IPFS{}, cmdArgs...)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = membrane.Start(ctx)
	require.NoError(t, err)

	// Create DefaultCell wrapper
	defaultCell := &system.DefaultCell{DefaultMembrane: membrane}

	// Wait for the process to complete
	err = defaultCell.Wait()
	require.NoError(t, err)
}

func TestDefaultCellKill(t *testing.T) {
	t.Parallel()

	cmdArgs := []string{"sleep", "10"}

	membrane, err := system.NewMembrane(system.IPFS{}, cmdArgs...)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = membrane.Start(ctx)
	require.NoError(t, err)

	// Create DefaultCell wrapper
	defaultCell := &system.DefaultCell{DefaultMembrane: membrane}

	// Wait a bit for the process to start
	time.Sleep(100 * time.Millisecond)

	// Kill the process
	err = defaultCell.Kill()
	require.NoError(t, err)

	// Wait for the process to be killed
	err = defaultCell.Wait()
	// The process was killed, so we expect an error
	require.Error(t, err)
	require.Contains(t, err.Error(), "signal: killed")
}

func TestDefaultCellGetPID(t *testing.T) {
	t.Parallel()

	cmdArgs := []string{"echo", "test"}

	membrane, err := system.NewMembrane(system.IPFS{}, cmdArgs...)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = membrane.Start(ctx)
	require.NoError(t, err)

	// Create DefaultCell wrapper
	defaultCell := &system.DefaultCell{DefaultMembrane: membrane}

	// Get the PID
	pid := defaultCell.GetPID()
	require.Greater(t, pid, 0)

	// Clean up
	defaultCell.Kill()
	defaultCell.Wait()
}

func TestMembraneAuthenticationFlow(t *testing.T) {
	t.Parallel()

	cmdArgs := []string{"echo", "test"}

	membrane, err := system.NewMembrane(system.IPFS{}, cmdArgs...)
	require.NoError(t, err)

	// Test that the membrane has the required fields for authentication
	require.NotNil(t, membrane.PrivateKey)
	require.NotEmpty(t, membrane.PeerID)
	require.Nil(t, membrane.Conn) // Should be nil initially
}

func TestMembraneFileDescriptors(t *testing.T) {
	t.Parallel()

	cmdArgs := []string{"echo", "test"}

	membrane, err := system.NewMembrane(system.IPFS{}, cmdArgs...)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = membrane.Start(ctx)
	require.NoError(t, err)

	// Check that file descriptors are not set up in test mode
	// In a real environment, these would be set up properly
	require.Len(t, membrane.Cmd.ExtraFiles, 0)

	// Clean up
	membrane.Cmd.Process.Kill()
}

func TestMembraneEnvironmentVariables(t *testing.T) {
	t.Parallel()

	cmdArgs := []string{"echo", "test"}

	membrane, err := system.NewMembrane(system.IPFS{}, cmdArgs...)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = membrane.Start(ctx)
	require.NoError(t, err)

	// Check that WW_ENV is set
	found := false
	for _, env := range membrane.Cmd.Env {
		if env == "WW_ENV=stdin,stdout,stderr" {
			found = true
			break
		}
	}
	require.True(t, found, "WW_ENV should be set in environment variables")

	// Clean up
	membrane.Cmd.Process.Kill()
}
