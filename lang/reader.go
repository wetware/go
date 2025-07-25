package lang

import (
	"fmt"
	"strings"

	"github.com/ipfs/boxo/path"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
)

// UnixPath represents a validated IPFS/IPLD path
type UnixPath struct {
	path.Path
}

// NewUnixPath creates a new UnixPath from a string, validating it using kubo's path.NewPath()
func NewUnixPath(pathStr string) (*UnixPath, error) {
	// First check if it starts with /ipfs or /ipld
	if !isValidPrefix(pathStr) {
		return nil, fmt.Errorf("invalid Unix path: must start with /ipfs/ or /ipld/")
	}

	// Validate using kubo's path.NewPath()
	p, err := path.NewPath(pathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid IPFS path: %w", err)
	}

	return &UnixPath{Path: p}, nil
}

// String returns the string representation of the path
func (up *UnixPath) String() string {
	return up.Path.String()
}

// ToBuiltinString converts the UnixPath to a builtin.String for use in the shell
func (up *UnixPath) ToBuiltinString() builtin.String {
	return builtin.String(up.Path.String())
}

// isValidPrefix checks if the path starts with /ipfs/ or /ipld/
func isValidPrefix(pathStr string) bool {
	return len(pathStr) >= 6 && (pathStr[:6] == "/ipfs/" || pathStr[:6] == "/ipld/")
}

// UnixPathReader implements a reader macro for Unix-style paths starting with /ipfs or /ipld
// This allows users to directly use IPFS/IPLD paths in the shell as literals.
// Examples:
//
//	/ipfs/QmHash123... -> returns UnixPath with validated path
//	/ipld/QmHash456... -> returns UnixPath with validated path
//	/invalid/path -> error: "invalid Unix path: must start with /ipfs/ or /ipld/"
func UnixPathReader() reader.Macro {
	return func(rd *reader.Reader, init rune) (core.Any, error) {
		// Read the path character by character until we hit whitespace or other delimiters
		var b strings.Builder
		b.WriteRune(init) // Start with the initial '/'

		for {
			r, err := rd.NextRune()
			if err != nil {
				// If we hit EOF, that's fine - we have a complete path
				break
			}

			// Stop at whitespace or other delimiters
			if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ')' || r == ']' || r == '}' {
				// Put the character back so the next reader can use it
				rd.Unread(r)
				break
			}

			b.WriteRune(r)
		}

		pathStr := b.String()

		// Create a validated UnixPath
		unixPath, err := NewUnixPath(pathStr)
		if err != nil {
			return nil, err
		}

		// Return the UnixPath as a core.Any
		return unixPath, nil
	}
}
