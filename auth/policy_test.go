package auth_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/system"
)

func TestConsoleConfig_Pattern(t *testing.T) {
	t.Parallel()
	// Test that ConsoleConfig follows the same pattern as IPFSConfig and ExecConfig

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Test ConsoleConfig instantiation
	config := system.ConsoleConfig{Writer: &buf}
	require.Equal(t, &buf, config.Writer, "Config should have the expected writer")

	// Test that New() returns a Console client
	console := config.New()
	require.NotNil(t, console, "New() should return a non-nil Console client")
	defer console.Release()

	// Test that we can add a reference
	consoleRef := console.AddRef()
	require.NotNil(t, consoleRef, "AddRef() should return a non-nil Console client")
	defer consoleRef.Release()
}

func TestConsoleConfig_ZeroValue(t *testing.T) {
	t.Parallel()
	// Test that zero value ConsoleConfig can be created
	var config system.ConsoleConfig
	require.Nil(t, config.Writer, "Zero value should have nil writer")

	// Test that New() still works (though it may not be useful with nil writer)
	console := config.New()
	require.NotNil(t, console, "New() should return a non-nil Console client even with zero value")
	defer console.Release()
}

func TestConsoleConfig_WithNilWriter(t *testing.T) {
	t.Parallel()
	// Test ConsoleConfig with nil writer
	config := system.ConsoleConfig{Writer: nil}

	// Test that New() returns a client
	console := config.New()
	require.NotNil(t, console, "New() should return a non-nil Console client even with nil writer")
	defer console.Release()
}
