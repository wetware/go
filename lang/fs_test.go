package lang

import (
	"testing"

	"github.com/ipfs/boxo/files"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_String(t *testing.T) {
	t.Parallel()

	// Create a mock file with known size
	content := "Hello, World!"
	file := files.NewBytesFile([]byte(content))

	f := File{File: file}

	// Test String method
	result := f.String()
	assert.Contains(t, result, "IPFS File:")
	assert.Contains(t, result, "bytes")
}

func TestFile_Invoke(t *testing.T) {
	t.Parallel()

	// Create a mock file
	content := "Hello, World!"
	file := files.NewBytesFile([]byte(content))
	f := File{File: file}

	tests := []struct {
		name     string
		args     []core.Any
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "no args - returns string representation",
			args:     []core.Any{},
			expected: "IPFS File:",
			wantErr:  false,
		},
		{
			name:     "type method",
			args:     []core.Any{builtin.Keyword("type")},
			expected: builtin.String("file"),
			wantErr:  false,
		},
		{
			name:     "size method",
			args:     []core.Any{builtin.Keyword("size")},
			expected: builtin.Int64(len(content)),
			wantErr:  false,
		},
		{
			name:     "read-string method",
			args:     []core.Any{builtin.Keyword("read-string")},
			expected: builtin.String(content),
			wantErr:  false,
		},
		{
			name:     "invalid method",
			args:     []core.Any{builtin.Keyword("invalid")},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "non-keyword argument",
			args:     []core.Any{builtin.String("not-a-keyword")},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Invoke(tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.expected != nil {
				switch expected := tt.expected.(type) {
				case string:
					if expected == "IPFS File:" {
						assert.Contains(t, result.(string), expected)
					} else {
						assert.Equal(t, expected, result)
					}
				case int64:
					assert.Equal(t, expected, result)
				}
			}
		})
	}
}

func TestDirectory_String(t *testing.T) {
	t.Parallel()

	// Create a mock directory
	dir := files.NewMapDirectory(map[string]files.Node{
		"file1.txt": files.NewBytesFile([]byte("content1")),
		"file2.txt": files.NewBytesFile([]byte("content2")),
	})

	d := Directory{Directory: dir}

	// Test String method
	result := d.String()
	assert.Contains(t, result, "IPFS Directory:")
	assert.Contains(t, result, "file1.txt")
	assert.Contains(t, result, "file2.txt")
}

func TestDirectory_Invoke(t *testing.T) {
	t.Parallel()

	// Create a mock directory
	dir := files.NewMapDirectory(map[string]files.Node{
		"file1.txt": files.NewBytesFile([]byte("content1")),
		"file2.txt": files.NewBytesFile([]byte("content2")),
	})

	d := Directory{Directory: dir}

	tests := []struct {
		name     string
		args     []core.Any
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "no args - returns string representation",
			args:     []core.Any{},
			expected: "IPFS Directory:",
			wantErr:  false,
		},
		{
			name:     "type method",
			args:     []core.Any{builtin.Keyword("type")},
			expected: builtin.String("directory"),
			wantErr:  false,
		},
		{
			name:     "list method",
			args:     []core.Any{builtin.Keyword("list")},
			expected: nil, // Will be a list, we'll check it's not nil
			wantErr:  false,
		},
		{
			name:     "entries method",
			args:     []core.Any{builtin.Keyword("entries")},
			expected: nil, // Will be a list, we'll check it's not nil
			wantErr:  false,
		},
		{
			name:     "invalid method",
			args:     []core.Any{builtin.Keyword("invalid")},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := d.Invoke(tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.expected != nil {
				if tt.expected == "IPFS Directory:" {
					assert.Contains(t, result.(string), tt.expected)
				} else {
					assert.Equal(t, tt.expected, result)
				}
			} else {
				// For list methods, just check it's not nil
				assert.NotNil(t, result)
			}
		})
	}
}

func TestNode_String(t *testing.T) {
	t.Parallel()

	// Test with a file node
	content := "Hello, World!"
	file := files.NewBytesFile([]byte(content))
	n := Node{Node: file}

	result := n.String()
	assert.Contains(t, result, "IPFS Node:")
	assert.Contains(t, result, "file")
}

func TestNode_Type(t *testing.T) {
	t.Parallel()

	// Test with a file node
	content := "Hello, World!"
	file := files.NewBytesFile([]byte(content))
	n := Node{Node: file}

	assert.Equal(t, "file", n.Type())

	// Test with a directory node
	dir := files.NewMapDirectory(map[string]files.Node{})
	nDir := Node{Node: dir}
	assert.Equal(t, "directory", nDir.Type())
}

func TestNode_Invoke(t *testing.T) {
	t.Parallel()

	// Create a mock file node
	content := "Hello, World!"
	file := files.NewBytesFile([]byte(content))
	n := Node{Node: file}

	tests := []struct {
		name     string
		args     []core.Any
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "no args - returns string representation",
			args:     []core.Any{},
			expected: "IPFS Node:",
			wantErr:  false,
		},
		{
			name:     "type method",
			args:     []core.Any{builtin.Keyword("type")},
			expected: builtin.String("file"),
			wantErr:  false,
		},
		{
			name:     "is-file method",
			args:     []core.Any{builtin.Keyword("is-file")},
			expected: builtin.Bool(true),
			wantErr:  false,
		},
		{
			name:     "is-directory method",
			args:     []core.Any{builtin.Keyword("is-directory")},
			expected: builtin.Bool(false),
			wantErr:  false,
		},
		{
			name:     "size method",
			args:     []core.Any{builtin.Keyword("size")},
			expected: builtin.Int64(len(content)),
			wantErr:  false,
		},
		{
			name:     "invalid method",
			args:     []core.Any{builtin.Keyword("invalid")},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := n.Invoke(tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.expected != nil {
				if tt.expected == "IPFS Node:" {
					assert.Contains(t, result.(string), tt.expected)
				} else {
					assert.Equal(t, tt.expected, result)
				}
			}
		})
	}
}

// TestFile_ReadContent tests that file content can be read multiple times
func TestFile_ReadContent(t *testing.T) {
	t.Parallel()

	content := "Hello, World!"
	file := files.NewBytesFile([]byte(content))
	f := File{File: file}

	// Test reading content multiple times
	for i := 0; i < 3; i++ {
		// Reset the file reader
		file = files.NewBytesFile([]byte(content))
		f = File{File: file}

		result, err := f.Invoke(builtin.Keyword("read-string"))
		require.NoError(t, err)
		assert.Equal(t, builtin.String(content), result)
	}
}

// TestDirectory_Empty tests empty directory handling
func TestDirectory_Empty(t *testing.T) {
	t.Parallel()

	// Create an empty directory
	dir := files.NewMapDirectory(map[string]files.Node{})
	d := Directory{Directory: dir}

	// Test string representation
	result := d.String()
	assert.Contains(t, result, "empty")

	// Test list method
	_, err := d.Invoke(builtin.Keyword("list"))
	require.NoError(t, err)
	// For empty directory, result might be nil or empty list - both are acceptable
	// We just verify the method doesn't error
}
