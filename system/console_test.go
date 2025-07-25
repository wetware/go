package system_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/system"
)

func TestConsoleConfig_New(t *testing.T) {
	t.Parallel()
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create ConsoleConfig
	config := system.ConsoleConfig{Writer: &buf}

	// Test that New() returns a Console client
	console := config.New()
	require.NotNil(t, console, "New() should return a non-nil Console client")

	// Test that we can release the client
	defer console.Release()

	// Verify the config has the expected writer
	require.Equal(t, &buf, config.Writer, "Config should have the expected writer")
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
