package lang

import (
	"fmt"
	"strings"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
)

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
		m := make(Map)
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

// Map represents an immutable map of keyword keys to any values
type Map map[builtin.Keyword]core.Any

// Get returns the value associated with the given key
func (m Map) Get(key builtin.Keyword) (core.Any, bool) {
	val, ok := m[key]
	return val, ok
}

// With returns a new Map with the given key-value pair added
func (m Map) With(key builtin.Keyword, val core.Any) Map {
	newEntries := make(map[builtin.Keyword]core.Any, len(m)+1)
	for k, v := range m {
		newEntries[k] = v
	}
	newEntries[key] = val
	return newEntries
}

// Without returns a new Map with the given key removed
func (m Map) Without(key builtin.Keyword) Map {
	if _, exists := m[key]; !exists {
		return m
	}

	newEntries := make(map[builtin.Keyword]core.Any, len(m)-1)
	for k, v := range m {
		if k != key {
			newEntries[k] = v
		}
	}
	return newEntries
}

// Len returns the number of entries in the map
func (m Map) Len() int {
	return len(m)
}

// SExpr implements core.SExpressable
func (m Map) SExpr() (string, error) {
	if len(m) == 0 {
		return "{}", nil
	}

	var s string
	for k, v := range m {
		if expr, ok := v.(core.SExpressable); ok {
			val, err := expr.SExpr()
			if err != nil {
				return "", err
			}
			s += fmt.Sprintf(" %s %s", k, val)
		} else {
			s += fmt.Sprintf(" %s %v", k, v)
		}
	}
	return "{" + s[1:] + "}", nil
}
