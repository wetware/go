# Wetware Language Extensions

This package provides language-level extensions for the Wetware shell, including custom reader macros and primitives.

## Unix Path Primitive

The `UnixPathReader` provides a custom reader macro that allows direct use of IPFS/IPLD paths as literals in the shell.

### Usage

Paths starting with `/ipfs/` or `/ipld/` are automatically parsed as string literals:

```lisp
;; Valid IPFS paths
/ipfs/QmHash123...     ; Returns string "/ipfs/QmHash123..."
/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa

;; Valid IPLD paths  
/ipld/QmHash456...     ; Returns string "/ipld/QmHash456..."

;; Invalid paths (will error)
/invalid/path          ; Error: "invalid Unix path: must start with /ipfs/ or /ipld/"
```

### Integration

The Unix path primitive is automatically available in the Wetware shell. It's implemented as a reader macro that triggers when encountering a `/` character at the beginning of a form.

### Examples

```lisp
;; Use IPFS paths directly in function calls
(ipfs.Cat /ipfs/QmHash123...)

;; Store paths in variables
(def my-path /ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa)

;; Use in lists and other data structures
(list /ipfs/QmHash1 /ipfs/QmHash2 /ipld/QmHash3)
```

### Implementation Details

The `UnixPathReader` function returns a `reader.Macro` that:

1. Reads characters until it encounters whitespace or delimiters
2. Validates that the path starts with `/ipfs/` or `/ipld/`
3. Returns the path as a `builtin.String` primitive

This allows for seamless integration with existing IPFS operations while providing a more natural syntax for working with IPFS/IPLD paths. 