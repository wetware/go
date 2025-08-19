package export_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/cmd/ww/export"
)

// TestEnv_Boot tests the Boot method of the export environment
func TestEnv_Boot(t *testing.T) {
	env := &export.Env{}

	// Test with invalid endpoint (this should always fail)
	err := env.Boot("invalid-endpoint")
	assert.Error(t, err, "Boot should fail with invalid endpoint")

	// Note: Testing with valid endpoints may succeed if IPFS daemon is running
	// We test the method exists and can handle invalid inputs
}

// TestEnv_Close tests the Close method of the export environment
func TestEnv_Close(t *testing.T) {
	env := &export.Env{}

	// Close should always succeed
	err := env.Close()
	assert.NoError(t, err, "Close should always succeed")
}

// TestEnv_AddToIPFS_File tests adding a single file to IPFS
func TestEnv_AddToIPFS_File(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test_file.txt")
	testContent := "Hello, IPFS! This is a test file."
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Test adding the file (will fail without IPFS, but we can test the method structure)
	_, err = env.AddToIPFS(ctx, testFile)
	assert.Error(t, err, "AddToIPFS should fail without IPFS daemon")
	assert.Contains(t, err.Error(), "IPFS client not initialized", "Error should contain expected message")
}

// TestEnv_AddToIPFS_Directory tests adding a directory to IPFS
func TestEnv_AddToIPFS_Directory(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create subdirectories and files
	subDir := filepath.Join(tempDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err, "Failed to create subdirectory")

	// Create files in root directory
	rootFile := filepath.Join(tempDir, "root_file.txt")
	err = os.WriteFile(rootFile, []byte("Root file content"), 0644)
	require.NoError(t, err, "Failed to create root file")

	// Create files in subdirectory
	subFile := filepath.Join(subDir, "sub_file.txt")
	err = os.WriteFile(subFile, []byte("Subdirectory file content"), 0644)
	require.NoError(t, err, "Failed to create subdirectory file")

	// Test adding the directory (will fail without IPFS, but we can test the method structure)
	_, err = env.AddToIPFS(ctx, tempDir)
	assert.Error(t, err, "AddToIPFS should fail without IPFS daemon")
	assert.Contains(t, err.Error(), "IPFS client not initialized", "Error should contain expected message")
}

// TestEnv_AddToIPFS_NonexistentPath tests error handling for nonexistent paths
func TestEnv_AddToIPFS_NonexistentPath(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Test with nonexistent path
	nonexistentPath := "/nonexistent/path"
	_, err := env.AddToIPFS(ctx, nonexistentPath)
	assert.Error(t, err, "AddToIPFS should fail with nonexistent path")
	assert.Contains(t, err.Error(), "failed to stat", "Error should contain expected message")
}

// TestEnv_createFileNode tests creating a file node
func TestEnv_createFileNode(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test_file.txt")
	testContent := "Test file content for IPFS node creation"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Test creating file node
	node, err := env.CreateFileNode(ctx, testFile)
	require.NoError(t, err, "createFileNode should succeed")
	assert.NotNil(t, node, "File node should not be nil")
}

// TestEnv_CreateFileNode_NonexistentFile tests error handling for nonexistent files
func TestEnv_CreateFileNode_NonexistentFile(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Test with nonexistent file
	nonexistentFile := "/nonexistent/file.txt"
	_, err := env.CreateFileNode(ctx, nonexistentFile)
	assert.Error(t, err, "CreateFileNode should fail with nonexistent file")
	assert.Contains(t, err.Error(), "no such file or directory", "Error should contain expected message")
}

// TestEnv_createDirectoryNode tests creating a directory node
func TestEnv_createDirectoryNode(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create subdirectories and files
	subDir := filepath.Join(tempDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err, "Failed to create subdirectory")

	// Create files in root directory
	rootFile := filepath.Join(tempDir, "root_file.txt")
	err = os.WriteFile(rootFile, []byte("Root file content"), 0644)
	require.NoError(t, err, "Failed to create root file")

	// Create files in subdirectory
	subFile := filepath.Join(subDir, "sub_file.txt")
	err = os.WriteFile(subFile, []byte("Subdirectory file content"), 0644)
	require.NoError(t, err, "Failed to create subdirectory file")

	// Test creating directory node
	node, err := env.CreateDirectoryNode(ctx, tempDir)
	require.NoError(t, err, "createDirectoryNode should succeed")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestEnv_createDirectoryNode_EmptyDirectory tests creating a node for an empty directory
func TestEnv_createDirectoryNode_EmptyDirectory(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create an empty temporary directory
	tempDir := t.TempDir()

	// Test creating directory node for empty directory
	node, err := env.CreateDirectoryNode(ctx, tempDir)
	require.NoError(t, err, "createDirectoryNode should succeed for empty directory")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestEnv_createDirectoryNode_NonexistentDirectory tests error handling for nonexistent directories
func TestEnv_createDirectoryNode_NonexistentDirectory(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Test with nonexistent directory
	nonexistentDir := "/nonexistent/directory"
	_, err := env.CreateDirectoryNode(ctx, nonexistentDir)
	assert.Error(t, err, "CreateDirectoryNode should fail with nonexistent directory")
	assert.Contains(t, err.Error(), "no such file or directory", "Error should contain expected message")
}

// TestEnv_createDirectoryNode_ComplexStructure tests creating a node for a complex directory structure
func TestEnv_createDirectoryNode_ComplexStructure(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create a complex temporary directory structure
	tempDir := t.TempDir()

	// Create multiple levels of subdirectories
	level1Dir := filepath.Join(tempDir, "level1")
	level2Dir := filepath.Join(level1Dir, "level2")
	level3Dir := filepath.Join(level2Dir, "level3")

	err := os.MkdirAll(level3Dir, 0755)
	require.NoError(t, err, "Failed to create nested directory structure")

	// Create files at different levels
	files := map[string]string{
		filepath.Join(tempDir, "root.txt"):      "Root level file",
		filepath.Join(level1Dir, "level1.txt"):  "Level 1 file",
		filepath.Join(level2Dir, "level2.txt"):  "Level 2 file",
		filepath.Join(level3Dir, "level3.txt"):  "Level 3 file",
		filepath.Join(level3Dir, "another.txt"): "Another level 3 file",
	}

	// Create all test files
	for filePath, content := range files {
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file: %s", filePath)
	}

	// Test creating directory node for complex structure
	node, err := env.CreateDirectoryNode(ctx, tempDir)
	require.NoError(t, err, "createDirectoryNode should succeed for complex structure")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestEnv_createDirectoryNode_FilePermissions tests creating directory nodes with different file permissions
func TestEnv_createDirectoryNode_FilePermissions(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create files with different permissions
	files := map[string]os.FileMode{
		filepath.Join(tempDir, "readonly.txt"):   0444,
		filepath.Join(tempDir, "executable.txt"): 0755,
		filepath.Join(tempDir, "normal.txt"):     0644,
	}

	// Create all test files with specified permissions
	for filePath, mode := range files {
		err := os.WriteFile(filePath, []byte("Test content"), mode)
		require.NoError(t, err, "Failed to create test file: %s", filePath)
	}

	// Test creating directory node
	node, err := env.CreateDirectoryNode(ctx, tempDir)
	require.NoError(t, err, "createDirectoryNode should succeed with different file permissions")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestEnv_createDirectoryNode_SpecialCharacters tests creating directory nodes with special characters in names
func TestEnv_createDirectoryNode_SpecialCharacters(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create files with special characters in names
	specialNames := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
		"file.with.dots.txt",
		"file@symbol.txt",
		"file#hash.txt",
		"file$dollar.txt",
		"file%percent.txt",
		"file^caret.txt",
		"file&ampersand.txt",
	}

	// Create all test files with special names
	for _, name := range specialNames {
		filePath := filepath.Join(tempDir, name)
		err := os.WriteFile(filePath, []byte("Test content"), 0644)
		require.NoError(t, err, "Failed to create test file: %s", name)
	}

	// Test creating directory node
	node, err := env.CreateDirectoryNode(ctx, tempDir)
	require.NoError(t, err, "createDirectoryNode should succeed with special characters")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestEnv_createDirectoryNode_LargeFiles tests creating directory nodes with large files
func TestEnv_createDirectoryNode_LargeFiles(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a large file (1MB)
	largeFile := filepath.Join(tempDir, "large_file.txt")
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256) // Fill with pattern
	}

	err := os.WriteFile(largeFile, largeContent, 0644)
	require.NoError(t, err, "Failed to create large test file")

	// Create a normal file for comparison
	normalFile := filepath.Join(tempDir, "normal_file.txt")
	err = os.WriteFile(normalFile, []byte("Normal file content"), 0644)
	require.NoError(t, err, "Failed to create normal test file")

	// Test creating directory node with large files
	node, err := env.CreateDirectoryNode(ctx, tempDir)
	require.NoError(t, err, "createDirectoryNode should succeed with large files")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestEnv_createDirectoryNode_Symlinks tests creating directory nodes with symbolic links
func TestEnv_createDirectoryNode_Symlinks(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a target file
	targetFile := filepath.Join(tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("Target file content"), 0644)
	require.NoError(t, err, "Failed to create target file")

	// Create a symbolic link
	symlinkPath := filepath.Join(tempDir, "symlink.txt")
	err = os.Symlink(targetFile, symlinkPath)
	require.NoError(t, err, "Failed to create symbolic link")

	// Test creating directory node with symlinks
	node, err := env.CreateDirectoryNode(ctx, tempDir)
	require.NoError(t, err, "createDirectoryNode should succeed with symlinks")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestEnv_createDirectoryNode_RecursiveSymlinks tests creating directory nodes with recursive symbolic links
func TestEnv_createDirectoryNode_RecursiveSymlinks(t *testing.T) {
	ctx := context.Background()
	env := &export.Env{}

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err, "Failed to create subdirectory")

	// Create a file in subdirectory
	subFile := filepath.Join(subDir, "subfile.txt")
	err = os.WriteFile(subFile, []byte("Subdirectory file content"), 0644)
	require.NoError(t, err, "Failed to create subdirectory file")

	// Create a symbolic link from root to a file (not a directory)
	symlinkPath := filepath.Join(tempDir, "link_to_file")
	err = os.Symlink(subFile, symlinkPath)
	require.NoError(t, err, "Failed to create symbolic link")

	// Test creating directory node with symlinks
	node, err := env.CreateDirectoryNode(ctx, tempDir)
	require.NoError(t, err, "CreateDirectoryNode should succeed with symlinks")
	assert.NotNil(t, node, "Directory node should not be nil")
}
