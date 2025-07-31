package lang

import (
	"fmt"
	"strings"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
)

// DotNotationReader implements a reader macro for dot notation like "ipfs.stat"
// This converts expressions like "ipfs.stat" into "(ipfs \"stat\")"
func DotNotationReader() reader.Macro {
	return func(rd *reader.Reader, init rune) (core.Any, error) {
		// Read the object name (before the dot)
		var objectName strings.Builder
		objectName.WriteRune(init) // Start with the initial character

		for {
			r, err := rd.NextRune()
			if err != nil {
				// If we hit EOF, this is just a symbol, not dot notation
				rd.Unread(init)
				return nil, fmt.Errorf("not dot notation")
			}

			if r == '.' {
				// Found the dot, now read the method name
				break
			}

			// Stop at whitespace or other delimiters
			if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ')' || r == ']' || r == '}' {
				// Put the character back and treat as regular symbol
				rd.Unread(r)
				rd.Unread(init)
				return nil, fmt.Errorf("not dot notation")
			}

			objectName.WriteRune(r)
		}

		// Read the method name (after the dot)
		var methodName strings.Builder
		for {
			r, err := rd.NextRune()
			if err != nil {
				// If we hit EOF, that's fine - we have a complete method name
				break
			}

			// Stop at whitespace or other delimiters
			if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ')' || r == ']' || r == '}' {
				// Put the character back so the next reader can use it
				rd.Unread(r)
				break
			}

			methodName.WriteRune(r)
		}

		// Create the function call form: (objectName "methodName")
		obj := builtin.Symbol(objectName.String())
		method := builtin.String(methodName.String())

		// Return a list that represents the function call
		return builtin.NewList(obj, method), nil
	}
}
