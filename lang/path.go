package lang

import (
	"fmt"

	"github.com/ipfs/boxo/path"
	"github.com/spy16/slurp/builtin"
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
