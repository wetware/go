package util

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIPFSEnv_Boot tests the Boot method
func TestIPFSEnv_Boot(t *testing.T) {
	env := IPFSEnv{}

	// Test with invalid endpoint (this should always fail)
	err := env.Boot("invalid-endpoint")
	assert.Error(t, err, "Boot should fail with invalid endpoint")

	// Note: Testing with valid endpoints may succeed if IPFS daemon is running
	// We test the method exists and can handle invalid inputs
}

// TestIPFSEnv_Close tests the Close method
func TestIPFSEnv_Close(t *testing.T) {
	env := IPFSEnv{}

	// Close should always succeed
	err := env.Close()
	assert.NoError(t, err, "Close should always succeed")
}

// TestIPFSEnv_GetIPFS tests the GetIPFS method
func TestIPFSEnv_GetIPFS(t *testing.T) {
	env := IPFSEnv{}

	// Test when IPFS is not initialized
	_, err := env.GetIPFS()
	assert.Error(t, err, "GetIPFS should fail when IPFS is not initialized")
	assert.Contains(t, err.Error(), "IPFS client not initialized", "Error should contain expected message")
}

// TestIPFSEnv_AddToIPFS_File tests adding a single file to IPFS
func TestIPFSEnv_AddToIPFS_File(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

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

// TestIPFSEnv_AddToIPFS_Directory tests adding a directory to IPFS
func TestIPFSEnv_AddToIPFS_Directory(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

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

// TestIPFSEnv_AddToIPFS_NonexistentPath tests error handling for nonexistent paths
func TestIPFSEnv_AddToIPFS_NonexistentPath(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Test with nonexistent path
	nonexistentPath := "/nonexistent/path"
	_, err := env.AddToIPFS(ctx, nonexistentPath)
	assert.Error(t, err, "AddToIPFS should fail with nonexistent path")
	assert.Contains(t, err.Error(), "failed to stat", "Error should contain expected message")
}

// TestIPFSEnv_CreateFileNode tests creating a file node
func TestIPFSEnv_CreateFileNode(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test_file.txt")
	testContent := "Test file content for IPFS node creation"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Test creating file node
	node, err := env.CreateFileNode(ctx, testFile)
	require.NoError(t, err, "CreateFileNode should succeed")
	assert.NotNil(t, node, "File node should not be nil")
}

// TestIPFSEnv_CreateFileNode_NonexistentFile tests error handling for nonexistent files
func TestIPFSEnv_CreateFileNode_NonexistentFile(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Test with nonexistent file
	nonexistentFile := "/nonexistent/file.txt"
	_, err := env.CreateFileNode(ctx, nonexistentFile)
	assert.Error(t, err, "CreateFileNode should fail with nonexistent file")
	assert.Contains(t, err.Error(), "no such file or directory", "Error should contain expected message")
}

// TestIPFSEnv_CreateDirectoryNode tests creating a directory node
func TestIPFSEnv_CreateDirectoryNode(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

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
	require.NoError(t, err, "CreateDirectoryNode should succeed")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestIPFSEnv_CreateDirectoryNode_EmptyDirectory tests creating a node for an empty directory
func TestIPFSEnv_CreateDirectoryNode_EmptyDirectory(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Create an empty temporary directory
	tempDir := t.TempDir()

	// Test creating directory node for empty directory
	node, err := env.CreateDirectoryNode(ctx, tempDir)
	require.NoError(t, err, "CreateDirectoryNode should succeed for empty directory")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestIPFSEnv_CreateDirectoryNode_NonexistentDirectory tests error handling for nonexistent directories
func TestIPFSEnv_CreateDirectoryNode_NonexistentDirectory(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Test with nonexistent directory
	nonexistentDir := "/nonexistent/directory"
	_, err := env.CreateDirectoryNode(ctx, nonexistentDir)
	assert.Error(t, err, "CreateDirectoryNode should fail with nonexistent directory")
	assert.Contains(t, err.Error(), "no such file or directory", "Error should contain expected message")
}

// TestIPFSEnv_CreateDirectoryNode_ComplexStructure tests creating a node for a complex directory structure
func TestIPFSEnv_CreateDirectoryNode_ComplexStructure(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

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
	require.NoError(t, err, "CreateDirectoryNode should succeed for complex structure")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestIPFSEnv_CreateDirectoryNode_FilePermissions tests creating directory nodes with different file permissions
func TestIPFSEnv_CreateDirectoryNode_FilePermissions(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

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
	require.NoError(t, err, "CreateDirectoryNode should succeed with different file permissions")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestIPFSEnv_CreateDirectoryNode_SpecialCharacters tests creating directory nodes with special characters in names
func TestIPFSEnv_CreateDirectoryNode_SpecialCharacters(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

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
	require.NoError(t, err, "CreateDirectoryNode should succeed with special characters")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestIPFSEnv_CreateDirectoryNode_LargeFiles tests creating directory nodes with large files
func TestIPFSEnv_CreateDirectoryNode_LargeFiles(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

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
	require.NoError(t, err, "CreateDirectoryNode should succeed with large files")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestIPFSEnv_CreateDirectoryNode_Symlinks tests creating directory nodes with symbolic links
func TestIPFSEnv_CreateDirectoryNode_Symlinks(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

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
	require.NoError(t, err, "CreateDirectoryNode should succeed with symlinks")
	assert.NotNil(t, node, "Directory node should not be nil")
}

// TestIPFSEnv_ImportFromIPFS tests the ImportFromIPFS method
func TestIPFSEnv_ImportFromIPFS(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Test with invalid IPFS path (will fail without IPFS, but we can test the method structure)
	invalidPath, _ := path.NewPath("/ipfs/invalid")

	// Test importing to current directory
	err := env.ImportFromIPFS(ctx, invalidPath, ".", false)
	assert.Error(t, err, "ImportFromIPFS should fail without IPFS daemon")
	assert.Contains(t, err.Error(), "IPFS client not initialized", "Error should contain expected message")
}

// TestIPFSEnv_ImportIPFSFile tests importing a single file from IPFS
func TestIPFSEnv_ImportIPFSFile(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a mock file node
	content := []byte("Test file content")
	fileNode := files.NewBytesFile(content)

	// Test importing file to directory
	err := env.ImportIPFSFile(ctx, fileNode, "/ipfs/test.txt", tempDir, false)
	assert.NoError(t, err, "ImportIPFSFile should succeed")

	// Verify file was created
	importedFile := filepath.Join(tempDir, "test.txt")
	require.FileExists(t, importedFile, "Imported file should exist")

	// Verify content
	importedContent, err := os.ReadFile(importedFile)
	require.NoError(t, err, "Should be able to read imported file")
	assert.Equal(t, content, importedContent, "Imported file content should match")
}

// TestIPFSEnv_ImportIPFSFile_ToSpecificPath tests importing a file to a specific path
func TestIPFSEnv_ImportIPFSFile_ToSpecificPath(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a mock file node
	content := []byte("Test file content")
	fileNode := files.NewBytesFile(content)

	// Test importing file to specific path
	specificPath := filepath.Join(tempDir, "custom_name.txt")
	err := env.ImportIPFSFile(ctx, fileNode, "/ipfs/test.txt", specificPath, false)
	assert.NoError(t, err, "ImportIPFSFile should succeed")

	// Verify file was created with custom name
	require.FileExists(t, specificPath, "Imported file should exist at custom path")

	// Verify content
	importedContent, err := os.ReadFile(specificPath)
	require.NoError(t, err, "Should be able to read imported file")
	assert.Equal(t, content, importedContent, "Imported file content should match")
}

// TestIPFSEnv_ImportIPFSFile_MakeExecutable tests importing a file with executable permissions
func TestIPFSEnv_ImportIPFSFile_MakeExecutable(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a mock file node
	content := []byte("#!/bin/bash\necho 'Hello World'")
	fileNode := files.NewBytesFile(content)

	// Test importing file with executable permissions
	importedFile := filepath.Join(tempDir, "script.sh")
	err := env.ImportIPFSFile(ctx, fileNode, "/ipfs/script.sh", importedFile, true)
	assert.NoError(t, err, "ImportIPFSFile should succeed")

	// Verify file was created
	require.FileExists(t, importedFile, "Imported file should exist")

	// Verify file is executable
	info, err := os.Stat(importedFile)
	require.NoError(t, err, "Should be able to stat imported file")
	assert.True(t, info.Mode()&0111 != 0, "Imported file should be executable")
}

// TestIPFSEnv_ImportIPFSDirectory tests importing a directory from IPFS
func TestIPFSEnv_ImportIPFSDirectory(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a mock directory node with files
	dirMap := map[string]files.Node{
		"file1.txt": files.NewBytesFile([]byte("File 1 content")),
		"file2.txt": files.NewBytesFile([]byte("File 2 content")),
		"subdir": files.NewMapDirectory(map[string]files.Node{
			"subfile.txt": files.NewBytesFile([]byte("Subdirectory file content")),
		}),
	}
	dirNode := files.NewMapDirectory(dirMap)

	// Test importing directory
	err := env.ImportIPFSDirectory(ctx, dirNode, "/ipfs/testdir", tempDir, false)
	assert.NoError(t, err, "ImportIPFSDirectory should succeed")

	// Verify files were created
	require.FileExists(t, filepath.Join(tempDir, "file1.txt"), "file1.txt should exist")
	require.FileExists(t, filepath.Join(tempDir, "file2.txt"), "file2.txt should exist")
	require.DirExists(t, filepath.Join(tempDir, "subdir"), "subdir should exist")
	require.FileExists(t, filepath.Join(tempDir, "subdir", "subfile.txt"), "subfile.txt should exist")
}

// TestIPFSEnv_ExtractIPFSDirectory tests recursively extracting an IPFS directory
func TestIPFSEnv_ExtractIPFSDirectory(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a mock directory node with nested structure
	dirMap := map[string]files.Node{
		"level1.txt": files.NewBytesFile([]byte("Level 1 file")),
		"nested": files.NewMapDirectory(map[string]files.Node{
			"level2.txt": files.NewBytesFile([]byte("Level 2 file")),
			"deep": files.NewMapDirectory(map[string]files.Node{
				"level3.txt": files.NewBytesFile([]byte("Level 3 file")),
			}),
		}),
	}
	dirNode := files.NewMapDirectory(dirMap)

	// Test extracting directory recursively
	err := env.ExtractIPFSDirectory(ctx, dirNode, tempDir, false)
	assert.NoError(t, err, "ExtractIPFSDirectory should succeed")

	// Verify nested structure was created
	require.FileExists(t, filepath.Join(tempDir, "level1.txt"), "level1.txt should exist")
	require.DirExists(t, filepath.Join(tempDir, "nested"), "nested directory should exist")
	require.FileExists(t, filepath.Join(tempDir, "nested", "level2.txt"), "level2.txt should exist")
	require.DirExists(t, filepath.Join(tempDir, "nested", "deep"), "deep directory should exist")
	require.FileExists(t, filepath.Join(tempDir, "nested", "deep", "level3.txt"), "level3.txt should exist")
}

// TestIPFSEnv_ExtractIPFSDirectory_WithExecutable tests extracting directory with executable files
func TestIPFSEnv_ExtractIPFSDirectory_WithExecutable(t *testing.T) {
	ctx := context.Background()
	env := IPFSEnv{}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a mock directory node with executable files
	dirMap := map[string]files.Node{
		"script.sh":  files.NewBytesFile([]byte("#!/bin/bash\necho 'Hello'")),
		"normal.txt": files.NewBytesFile([]byte("Normal file")),
	}
	dirNode := files.NewMapDirectory(dirMap)

	// Test extracting directory with executable permissions
	err := env.ExtractIPFSDirectory(ctx, dirNode, tempDir, true)
	assert.NoError(t, err, "ExtractIPFSDirectory should succeed")

	// Verify files were created
	scriptPath := filepath.Join(tempDir, "script.sh")
	normalPath := filepath.Join(tempDir, "normal.txt")
	require.FileExists(t, scriptPath, "script.sh should exist")
	require.FileExists(t, normalPath, "normal.txt should exist")

	// Verify script is executable
	scriptInfo, err := os.Stat(scriptPath)
	require.NoError(t, err, "Should be able to stat script file")
	assert.True(t, scriptInfo.Mode()&0111 != 0, "Script file should be executable")

	// Note: We only check that the script file is executable, not that normal files are non-executable
	// as file creation permissions can vary by system
}

// TestIPFSEnv_isDirectory tests the isDirectory helper function
func TestIPFSEnv_isDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test with directory
	assert.True(t, isDirectory(tempDir), "isDirectory should return true for directory")

	// Test with file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err, "Failed to create test file")
	assert.False(t, isDirectory(testFile), "isDirectory should return false for file")

	// Test with nonexistent path
	assert.False(t, isDirectory("/nonexistent/path"), "isDirectory should return false for nonexistent path")
}
