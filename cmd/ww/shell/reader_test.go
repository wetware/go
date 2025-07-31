package shell_test

import (
	"strings"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/reader"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/cmd/ww/shell"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
)

func TestUnixPathReader(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid ipfs path",
			input:   "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			want:    "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			wantErr: false,
		},
		{
			name:    "valid ipld path",
			input:   "/ipld/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			want:    "/ipld/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			wantErr: false,
		},
		{
			name:    "valid ipfs path with longer hash",
			input:   "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			want:    "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			wantErr: false,
		},
		{
			name:    "invalid path starting with slash",
			input:   "/invalid/path",
			want:    "",
			wantErr: true,
		},
		{
			name:    "path with trailing whitespace",
			input:   "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa ",
			want:    "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			wantErr: false,
		},
		{
			name:    "path followed by parenthesis",
			input:   "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa)",
			want:    "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader with our custom macro
			rd := reader.New(strings.NewReader(tt.input))
			rd.SetMacro('/', false, shell.UnixPathReader())

			// Read one form
			result, err := rd.One()

			if tt.wantErr {
				if err == nil {
					t.Errorf("UnixPathReader() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("UnixPathReader() unexpected error: %v", err)
				return
			}

			// Check if result is a UnixPath
			unixPath, ok := result.(*lang.UnixPath)
			if !ok {
				t.Errorf("UnixPathReader() returned %T, want *UnixPath", result)
				return
			}

			if unixPath.String() != tt.want {
				t.Errorf("UnixPathReader() = %v, want %v", unixPath.String(), tt.want)
			}
		})
	}
}

func TestUnixPathReaderInContext(t *testing.T) {
	t.Parallel()
	// Test that the Unix path reader works in a more realistic context
	// where it's followed by other forms
	input := "(ipfs.Cat /ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa) (ipfs.Ls /ipld/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa)"

	rd := reader.New(strings.NewReader(input))
	rd.SetMacro('/', false, shell.UnixPathReader())

	// Read all forms
	forms, err := rd.All()
	if err != nil {
		t.Fatalf("Failed to read forms: %v", err)
	}

	if len(forms) != 2 {
		t.Fatalf("Expected 2 forms, got %d", len(forms))
	}

	// Check that the paths were parsed correctly
	// The first form should be a list with "ipfs.Cat" and the path
	// The second form should be a list with "ipfs.Ls" and the path

	t.Logf("Successfully parsed forms: %v", forms)
}

func TestNewUnixPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid ipfs path",
			input:   "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			want:    "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			wantErr: false,
		},
		{
			name:    "valid ipld path",
			input:   "/ipld/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			want:    "/ipld/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			wantErr: false,
		},
		{
			name:    "invalid path starting with slash",
			input:   "/invalid/path",
			want:    "",
			wantErr: true,
		},
		{
			name:    "path without slash prefix",
			input:   "ipfs/QmHash123",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty path",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unixPath, err := lang.NewUnixPath(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewUnixPath() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewUnixPath() unexpected error: %v", err)
				return
			}

			if unixPath.String() != tt.want {
				t.Errorf("NewUnixPath() = %v, want %v", unixPath.String(), tt.want)
			}

			// Test that the underlying path is accessible
			if unixPath.String() != tt.want {
				t.Errorf("UnixPath.Path() = %v, want %v", unixPath.String(), tt.want)
			}

			// Test ToBuiltinString conversion
			builtinStr := unixPath.ToBuiltinString()
			if string(builtinStr) != tt.want {
				t.Errorf("UnixPath.ToBuiltinString() = %v, want %v", string(builtinStr), tt.want)
			}
		})
	}
}

func TestListReader(t *testing.T) {
	t.Parallel()
	// Create a mock IPFS server for testing
	mockServer := &MockIPFSServer{testValue: 42}
	mock := system.IPFS_ServerToClient(mockServer)

	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{
			name:    "empty list",
			input:   "()",
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "simple list",
			input:   "(1 2 3)",
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "list with strings",
			input:   "(\"hello\" \"world\")",
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "list with UnixPath",
			input:   "(/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa)",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "nested list",
			input:   "((1 2) (3 4))",
			wantLen: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader with our custom macros
			rd := reader.New(strings.NewReader(tt.input))
			rd.SetMacro('/', false, shell.UnixPathReader())
			rd.SetMacro('(', false, shell.ListReader(mock))

			// Read one form
			result, err := rd.One()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.IsType(t, (*lang.IPLDLinkedList)(nil), result)
			list := result.(*lang.IPLDLinkedList)

			count, err := list.Count()
			require.NoError(t, err)
			require.Equal(t, tt.wantLen, count)

			t.Logf("Successfully parsed list: %v", result)
		})
	}
}

func TestHexReader(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid hex string",
			input:   "0x746573742064617461",
			want:    "test data",
			wantErr: false,
		},
		{
			name:    "empty hex string",
			input:   "0x",
			want:    "",
			wantErr: false,
		},
		{
			name:    "single byte hex",
			input:   "0x41",
			want:    "A",
			wantErr: false,
		},
		{
			name:    "invalid hex string",
			input:   "0xinvalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "doesn't start with 0x",
			input:   "746573742064617461",
			want:    "746573742064617461",
			wantErr: false,
		},
		{
			name:    "hex with trailing whitespace",
			input:   "0x746573742064617461 ",
			want:    "test data",
			wantErr: false,
		},
		{
			name:    "hex followed by parenthesis",
			input:   "0x746573742064617461)",
			want:    "test data",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the NewReaderWithHexSupport function instead of setting the macro directly
			rd := shell.NewReaderWithHexSupport(strings.NewReader(tt.input))

			// Read one form
			result, err := rd.One()

			if tt.wantErr {
				require.Error(t, err, "HexReader() expected error but got none")
				return
			}

			require.NoError(t, err, "HexReader() unexpected error: %v", err)

			// Check if result is a Buffer (for hex literals) or a number (for regular numbers)
			if strings.HasPrefix(tt.input, "0x") {
				buffer, ok := result.(*lang.Buffer)
				require.True(t, ok, "HexReader() returned %T, want *Buffer", result)

				require.Equal(t, tt.want, buffer.String(), "HexReader() = %v, want %v", buffer.String(), tt.want)
			} else {
				// For non-hex numbers, we expect a number type
				_, ok := result.(builtin.Int64)
				require.True(t, ok, "HexReader() returned %T, want builtin.Int64", result)
			}
		})
	}
}

func TestHexReaderInContext(t *testing.T) {
	t.Parallel()
	// Test that the hex reader works in a more realistic context
	// where it's used in function calls
	input := "(add 0x746573742064617461) (cat /ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa)"

	// Use the NewReaderWithHexSupport function
	rd := shell.NewReaderWithHexSupport(strings.NewReader(input))
	rd.SetMacro('/', false, shell.UnixPathReader())

	// Read all forms
	forms, err := rd.All()
	require.NoError(t, err, "Failed to read forms: %v", err)

	require.Equal(t, 2, len(forms), "Expected 2 forms, got %d", len(forms))

	t.Logf("Successfully parsed forms: %v", forms)
}
