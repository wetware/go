package lang

import (
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	"capnproto.org/go/capnp/v3/exp/bufferpool"
	"github.com/ipfs/boxo/path"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
	"github.com/wetware/go/system"
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

// HexReader implements a reader macro for hex-encoded bytes starting with "0x"
// This allows users to directly use hex literals in the shell.
// Examples:
//
//	0x746573742064617461 -> returns Buffer with decoded bytes ("test data")
//	0x -> returns empty Buffer
//	0xinvalid -> error: "invalid hex string"
func HexReader() reader.Macro {
	return func(rd *reader.Reader, init rune) (core.Any, error) {
		// Check if the next character is 'x' to confirm this is a hex literal
		nextRune, err := rd.NextRune()
		if err != nil {
			// If we hit EOF, this is just "0" not a hex literal
			rd.Unread(init)
			return nil, fmt.Errorf("unexpected EOF")
		}

		if nextRune != 'x' {
			// Not a hex literal, put both characters back and let the default reader handle it
			rd.Unread(nextRune)
			rd.Unread(init)
			return nil, fmt.Errorf("not a hex literal")
		}

		// Read the rest of the hex string character by character until we hit whitespace or other delimiters
		var b strings.Builder
		b.WriteRune(init)     // Start with the initial '0'
		b.WriteRune(nextRune) // Add the 'x'

		for {
			r, err := rd.NextRune()
			if err != nil {
				// If we hit EOF, that's fine - we have a complete hex string
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

		hexStr := b.String()

		// Handle empty hex string (just "0x")
		if hexStr == "0x" {
			return &Buffer{}, nil
		}

		// Decode the hex string (skip the "0x" prefix)
		hexData := hexStr[2:]
		data, err := hex.DecodeString(hexData)
		if err != nil {
			return nil, fmt.Errorf("invalid hex string: %w", err)
		}

		// Create a buffer with the decoded data
		buf := bufferpool.Default.Get(len(data))
		copy(buf, data)

		return &Buffer{Mem: buf}, nil
	}
}

// ListReader creates a custom list reader macro that can access the IPFS session
// This allows for enhanced list processing with IPFS capabilities
func ListReader(ipfs system.IPFS) reader.Macro {
	return func(rd *reader.Reader, init rune) (core.Any, error) {
		const listEnd = ')'

		forms := make([]core.Any, 0, 32) // pre-allocate to improve performance on small lists
		if err := rd.Container(listEnd, "list", func(val core.Any) error {
			forms = append(forms, val)
			return nil
		}); err != nil {
			return nil, err
		}

		// For now, we'll just return a regular list like the default reader
		// In the future, this could be enhanced to handle IPFS-specific list operations
		// such as:
		// - Auto-resolving IPFS paths in lists
		// - Special handling for IPFS commands
		// - Batch operations on IPFS objects

		return builtin.NewList(forms...), nil
	}
}

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

// NewReaderWithHexSupport creates a new reader with support for hex literals and map literals
// This allows users to use hex literals like 0x746573742064617461 and map literals like {:a 1 :b 2}
func NewReaderWithHexSupport(r io.Reader) *reader.Reader {
	// Create a custom number reader that can handle both hex literals and regular numbers
	customNumReader := func(rd *reader.Reader, init rune) (core.Any, error) {
		// Check if this is a hex literal (starts with "0x")
		if init == '0' {
			// Check if the next character is 'x' to confirm this is a hex literal
			nextRune, err := rd.NextRune()
			if err != nil {
				// If we hit EOF, this is just "0" not a hex literal
				rd.Unread(init)
				return nil, fmt.Errorf("unexpected EOF")
			}

			if nextRune == 'x' {
				// This is a hex literal, parse it
				return parseHexLiteral(rd, init, nextRune)
			} else {
				// Not a hex literal, put the character back and parse as regular number
				rd.Unread(nextRune)
			}
		}

		// Parse as a regular number
		return parseRegularNumber(rd, init)
	}

	// Create the reader with our custom number reader
	rd := reader.New(r, reader.WithNumReader(customNumReader))

	// Set up the map reader macro for '{' character
	rd.SetMacro('{', false, MapReader())

	return rd
}

// parseHexLiteral parses a hex literal starting with "0x"
func parseHexLiteral(rd *reader.Reader, init rune, nextRune rune) (core.Any, error) {
	// Read the rest of the hex string character by character until we hit whitespace or other delimiters
	var b strings.Builder
	b.WriteRune(init)     // Start with the initial '0'
	b.WriteRune(nextRune) // Add the 'x'

	for {
		r, err := rd.NextRune()
		if err != nil {
			// If we hit EOF, that's fine - we have a complete hex string
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

	hexStr := b.String()

	// Handle empty hex string (just "0x")
	if hexStr == "0x" {
		return &Buffer{}, nil
	}

	// Decode the hex string (skip the "0x" prefix)
	hexData := hexStr[2:]
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}

	buf := bufferpool.Default.Get(len(data))
	copy(buf, data)

	// Create a buffer with the decoded data
	buffer := &Buffer{Mem: buf}

	return buffer, nil
}

// parseRegularNumber parses a regular number
func parseRegularNumber(rd *reader.Reader, init rune) (core.Any, error) {
	var b strings.Builder
	b.WriteRune(init)

	for {
		r, err := rd.NextRune()
		if err != nil {
			break
		}

		// Stop at whitespace or other delimiters
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ')' || r == ']' || r == '}' {
			rd.Unread(r)
			break
		}

		// Allow digits, decimal point, and minus sign
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			b.WriteRune(r)
		} else {
			// Put the character back and stop
			rd.Unread(r)
			break
		}
	}

	numStr := b.String()

	// Try to parse as an integer first
	if i, err := strconv.ParseInt(numStr, 10, 64); err == nil {
		return builtin.Int64(i), nil
	}

	// Try to parse as a float
	if f, err := strconv.ParseFloat(numStr, 64); err == nil {
		return builtin.Float64(f), nil
	}

	// If all else fails, return as a string
	return builtin.String(numStr), nil
}
