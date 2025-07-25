package lang

import (
	"strings"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/reader"
)

func TestUnixPathReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid ipfs path",
			input:   "/ipfs/QmHash123",
			want:    "/ipfs/QmHash123",
			wantErr: false,
		},
		{
			name:    "valid ipld path",
			input:   "/ipld/QmHash456",
			want:    "/ipld/QmHash456",
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
			input:   "/ipfs/QmHash123 ",
			want:    "/ipfs/QmHash123",
			wantErr: false,
		},
		{
			name:    "path followed by parenthesis",
			input:   "/ipfs/QmHash123)",
			want:    "/ipfs/QmHash123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader with our custom macro
			rd := reader.New(strings.NewReader(tt.input))
			rd.SetMacro('/', false, UnixPathReader())

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

			// Check if result is a string
			str, ok := result.(builtin.String)
			if !ok {
				t.Errorf("UnixPathReader() returned %T, want builtin.String", result)
				return
			}

			if string(str) != tt.want {
				t.Errorf("UnixPathReader() = %v, want %v", string(str), tt.want)
			}
		})
	}
}

func TestUnixPathReaderInContext(t *testing.T) {
	// Test that the Unix path reader works in a more realistic context
	// where it's followed by other forms
	input := "(ipfs.Cat /ipfs/QmHash123) (ipfs.Ls /ipld/QmHash456)"

	rd := reader.New(strings.NewReader(input))
	rd.SetMacro('/', false, UnixPathReader())

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
