package lang

import (
	"fmt"
	"strings"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
)

// UnixPathReader implements a reader macro for Unix-style paths starting with /ipfs or /ipld
// This allows users to directly use IPFS/IPLD paths in the shell as literals.
// Examples:
//
//	/ipfs/QmHash123... -> returns string "/ipfs/QmHash123..."
//	/ipld/QmHash456... -> returns string "/ipld/QmHash456..."
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

		path := b.String()

		// Validate that the path starts with /ipfs or /ipld
		if !strings.HasPrefix(path, "/ipfs/") && !strings.HasPrefix(path, "/ipld/") {
			return nil, fmt.Errorf("invalid Unix path: must start with /ipfs/ or /ipld/")
		}

		// Return the path as a string primitive
		return builtin.String(path), nil
	}
}
