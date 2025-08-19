package main

import (
	"strings"
	"testing"

	"github.com/spy16/slurp/reader"
)

func TestIPFSPathReader(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "valid IPFS path",
			input:    "/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi/example",
			expected: "Path",
			wantErr:  false,
		},
		{
			name:     "valid IPNS path",
			input:    "/ipns/example.com/file",
			expected: "Path",
			wantErr:  false,
		},
		{
			name:     "invalid path starting with slash",
			input:    "/foo/bar",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "single slash",
			input:    "/",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "path with spaces",
			input:    "/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi/example file",
			expected: "Path",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader with our test input
			rd := reader.New(strings.NewReader(tt.input))

			// Call the IPFS path reader
			result, err := IPFSPathReader(rd, '/')

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("IPFSPathReader() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("IPFSPathReader() unexpected error: %v", err)
				return
			}

			// Check result type
			if tt.expected == "Path" {
				if _, ok := result.(Path); !ok {
					t.Errorf("IPFSPathReader() expected Path type, got %T", result)
				}
			}
		})
	}
}

func TestIPFSPathReaderWithValidPaths(t *testing.T) {
	t.Parallel()
	validPaths := []string{
		"/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
		"/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi/example",
		"/ipfs/bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi/example/file.txt",
		"/ipns/example.com",
		"/ipns/example.com/file",
		"/ipns/example.com/deep/path/file.txt",
	}

	for _, pathStr := range validPaths {
		t.Run("valid_"+pathStr, func(t *testing.T) {
			rd := reader.New(strings.NewReader(pathStr))

			result, err := IPFSPathReader(rd, '/')
			if err != nil {
				t.Errorf("IPFSPathReader() failed for valid path %s: %v", pathStr, err)
				return
			}

			pathObj, ok := result.(Path)
			if !ok {
				t.Errorf("IPFSPathReader() returned wrong type for %s: %T", pathStr, result)
				return
			}

			// Verify the path string matches
			if pathObj.String() != pathStr {
				t.Errorf("IPFSPathReader() returned path %s, expected %s", pathObj.String(), pathStr)
			}
		})
	}
}

func TestIPFSPathReaderWithInvalidPaths(t *testing.T) {
	t.Parallel()
	invalidPaths := []string{
		"/foo",
		"/bar/baz",
		"/notipfs/path",
		"/notipns/domain",
		"/random/stuff",
	}

	for _, pathStr := range invalidPaths {
		t.Run("invalid_"+pathStr, func(t *testing.T) {
			rd := reader.New(strings.NewReader(pathStr))

			result, err := IPFSPathReader(rd, '/')
			if err == nil {
				t.Errorf("IPFSPathReader() should have failed for invalid path %s, got result: %v", pathStr, result)
			}

			// Check that it's a reader error
			if _, ok := err.(*reader.Error); !ok {
				t.Errorf("IPFSPathReader() should return reader.Error for invalid path %s, got: %T", pathStr, err)
			}
		})
	}
}

func TestIPFSPathReaderEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty after slash",
			input:   "/",
			wantErr: true,
		},
		{
			name:    "just ipfs",
			input:   "/ipfs",
			wantErr: true,
		},
		{
			name:    "just ipns",
			input:   "/ipns",
			wantErr: true,
		},
		{
			name:    "ipfs with trailing slash",
			input:   "/ipfs/",
			wantErr: true,
		},
		{
			name:    "ipns with trailing slash",
			input:   "/ipns/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := reader.New(strings.NewReader(tt.input))

			result, err := IPFSPathReader(rd, '/')
			if tt.wantErr {
				if err == nil {
					t.Errorf("IPFSPathReader() expected error for %s but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("IPFSPathReader() unexpected error for %s: %v", tt.input, err)
				}
			}

			// If no error, result should be a Path
			if err == nil {
				if _, ok := result.(Path); !ok {
					t.Errorf("IPFSPathReader() returned wrong type for %s: %T", tt.input, result)
				}
			}
		})
	}
}
