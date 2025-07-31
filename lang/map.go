package lang

import (
	"fmt"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

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
