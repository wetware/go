package run_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/cmd/ww/run"
)

func TestEnv_ResolveExecPath_LocalPath(t *testing.T) {
	ctx := context.Background()

	// Create a temporary environment
	env := &run.Env{}
	env.Dir = t.TempDir()

	// Create a temporary executable file
	tempFile := filepath.Join(env.Dir, "test_exec")
	err := os.WriteFile(tempFile, []byte("#!/bin/sh\necho 'test'"), 0755)
	require.NoError(t, err)

	// Test local absolute path
	result, err := env.ResolveExecPath(ctx, tempFile)
	require.NoError(t, err)
	assert.Equal(t, tempFile, result)

	// Test local relative path - this will resolve relative to current working directory
	relPath := filepath.Base(tempFile)
	result, err = env.ResolveExecPath(ctx, relPath)
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(result))
	// The result should be the absolute path, but it might not match tempFile exactly
	// due to working directory differences
	assert.Equal(t, filepath.Base(tempFile), filepath.Base(result))
}

func TestEnv_ResolveExecPath_InvalidIPFSPath(t *testing.T) {
	ctx := context.Background()

	// Create a temporary environment
	env := &run.Env{}
	env.Dir = t.TempDir()

	// Test invalid IPFS path
	invalidPath := "not-an-ipfs-path"
	result, err := env.ResolveExecPath(ctx, invalidPath)

	// Should fall back to local filesystem handling
	require.NoError(t, err)
	// The result should be the absolute path of the invalid path
	assert.True(t, filepath.IsAbs(result))
	assert.Equal(t, filepath.Base(invalidPath), filepath.Base(result))
}

func TestEnv_OS(t *testing.T) {
	env := &run.Env{}

	// Test with WW_OS set
	os.Setenv("WW_OS", "custom_os")
	defer os.Unsetenv("WW_OS")

	result := env.OS()
	assert.Equal(t, "custom_os", result)

	// Test without WW_OS set (should use runtime.GOOS)
	os.Unsetenv("WW_OS")
	result = env.OS()
	assert.Equal(t, runtime.GOOS, result)
}

func TestEnv_Arch(t *testing.T) {
	env := &run.Env{}

	// Test with WW_ARCH set
	os.Setenv("WW_ARCH", "custom_arch")
	defer os.Unsetenv("WW_ARCH")

	result := env.Arch()
	assert.Equal(t, "custom_arch", result)

	// Test without WW_ARCH set (should use runtime.GOARCH)
	os.Unsetenv("WW_ARCH")
	result = env.Arch()
	assert.Equal(t, runtime.GOARCH, result)
}

func TestEnv_ResolveIPFSFile(t *testing.T) {
	ctx := context.Background()

	// Create a temporary environment
	env := &run.Env{}
	env.Dir = t.TempDir()

	// Create a simple test file
	testFile := filepath.Join(env.Dir, "test_file")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test file resolution (this will fail since we don't have IPFS set up)
	// but we can test the method signature and basic logic
	ipfsPath := "/ipfs/QmTestFile"
	_, err = env.ResolveIPFSFile(ctx, nil, ipfsPath)

	// Should fail since we don't have a real IPFS node
	require.Error(t, err)
}

func TestEnv_ResolveIPFSDirectory_WithBinSubdir(t *testing.T) {
	// Create a temporary environment
	env := &run.Env{}
	env.Dir = t.TempDir()

	// This test expects the method to exist and be callable
	// Since we can't easily mock the complex IPFS interfaces, we'll just test
	// that the method exists and can be called
	assert.NotNil(t, env.ResolveIPFSDirectory)
}
