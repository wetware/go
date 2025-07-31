package shell

import (
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	"capnproto.org/go/capnp/v3/exp/bufferpool"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
	"github.com/spy16/slurp/repl"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
)

// readerFactory creates a custom reader factory for the shell
func readerFactory(ipfs system.IPFS) repl.ReaderFactoryFunc {
	return func(r io.Reader) *reader.Reader {
		// Create a reader with hex support
		rd := NewReaderWithHexSupport(r)

		rd.SetMacro('/', false, UnixPathReader())
		rd.SetMacro('(', false, ListReader(ipfs))
		rd.SetMacro('{', false, MapReader())

		return rd
	}
}

// NewReaderWithHexSupport creates a new reader with support for hex literals
// This allows users to use hex literals like 0x746573742064617461
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
		return &lang.Buffer{}, nil
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
	buffer := &lang.Buffer{Mem: buf}

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
		unixPath, err := lang.NewUnixPath(pathStr)
		if err != nil {
			return nil, err
		}

		// Return the UnixPath as a core.Any
		return unixPath, nil
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

// ListReader creates a custom list reader macro that can access the IPFS session
// This allows for enhanced list processing with IPFS capabilities
func ListReader(ipfs system.IPFS) reader.Macro {
	return func(rd *reader.Reader, init rune) (core.Any, error) {
		const listEnd = ')'

		// Collect all values in a slice
		var values []core.Any
		if err := rd.Container(listEnd, "list", func(val core.Any) (err error) {
			values = append(values, val)
			return nil
		}); err != nil {
			return nil, err
		}

		// Create an immutable/persistent linked list using builtin.LinkedList
		// In a more sophisticated implementation, this would serialize the values to IPFS DAG
		// and create a proper linked list structure with minimal data linking to head and next nodes
		if len(values) == 0 {
			// Empty list - create IPLD list for consistency
			ipldList, err := lang.NewIPLDLinkedList(ipfs)
			if err != nil {
				// Fallback to builtin.LinkedList if IPLD creation fails
				return builtin.NewList(), nil
			}
			return ipldList, nil
		}

		// Create a proper IPLD-based immutable/persistent linked list
		// This creates a DAG structure where each node is stored separately in IPFS
		// with minimal data linking to head and next nodes
		ipldList, err := lang.NewIPLDLinkedList(ipfs, values...)
		if err != nil {
			// Fallback to builtin.LinkedList if IPLD creation fails
			return builtin.NewList(values...), nil
		}

		return ipldList, nil
	}
}

// MapReader implements a reader macro for map literals using curly braces
// This allows users to create maps like {:a 1 :b 2 :c 3}
// The syntax is {key1 val1 key2 val2 ...} where keys should be keywords
func MapReader() reader.Macro {
	return func(rd *reader.Reader, init rune) (core.Any, error) {
		const mapEnd = '}'

		// Read all forms within the map
		forms := make([]core.Any, 0, 32)
		if err := rd.Container(mapEnd, "map", func(val core.Any) error {
			forms = append(forms, val)
			return nil
		}); err != nil {
			return nil, err
		}

		// Check that we have an even number of forms (key-value pairs)
		if len(forms)%2 != 0 {
			return nil, fmt.Errorf("map literal must have even number of forms (key-value pairs), got %d", len(forms))
		}

		// Create the map
		m := make(lang.Map)
		for i := 0; i < len(forms); i += 2 {
			key := forms[i]
			value := forms[i+1]

			// Convert key to keyword if it's not already
			var keyword builtin.Keyword
			switch k := key.(type) {
			case builtin.Keyword:
				keyword = k
			case builtin.String:
				// Handle case where string might have trailing colon
				str := strings.TrimSuffix(string(k), ":")
				keyword = builtin.Keyword(str)
			case builtin.Symbol:
				// Handle case where symbol might have trailing colon
				sym := strings.TrimSuffix(string(k), ":")
				keyword = builtin.Keyword(sym)
			default:
				return nil, fmt.Errorf("map key must be keyword, string, or symbol, got %T", key)
			}

			// Check for duplicate keys
			if _, exists := m[keyword]; exists {
				return nil, fmt.Errorf("duplicate key in map: %v", keyword)
			}

			m[keyword] = value
		}

		return m, nil
	}
}
