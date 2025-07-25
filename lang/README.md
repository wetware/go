# Wetware Language Extensions

This package provides language-level extensions for the Wetware shell, including custom reader macros and primitives.

## Unix Path Primitive

The `UnixPathReader` provides a custom reader macro that allows direct use of IPFS/IPLD paths as literals in the shell. Paths are validated using kubo's `path.NewPath()` function to ensure they are valid IPFS paths.

### UnixPath Type

The `UnixPath` type represents a validated IPFS/IPLD path with the following features:

- **Validation**: Uses kubo's `path.NewPath()` to validate path format and CID encoding
- **Type Safety**: Provides a strongly-typed wrapper around IPFS paths
- **Conversion**: Can be converted to `builtin.String` for use in shell operations
- **Access**: Provides access to the underlying kubo `path.Path` object

### Usage

Paths starting with `/ipfs/` or `/ipld/` are automatically parsed as `UnixPath` objects:

```lisp
;; Valid IPFS paths (returns UnixPath objects)
/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa
/ipld/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa

;; Invalid paths (will error)
/invalid/path          ; Error: "invalid Unix path: must start with /ipfs/ or /ipld/"
/ipfs/invalid-cid      ; Error: "invalid IPFS path: invalid cid: selected encoding not supported"
```

### Integration

The Unix path primitive is automatically available in the Wetware shell. It's implemented as a reader macro that triggers when encountering a `/` character at the beginning of a form.

### Examples

```lisp
;; Use IPFS paths directly in function calls
(ipfs.Cat /ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa)

;; Store paths in variables (UnixPath objects)
(def my-path /ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa)

;; Use in lists and other data structures
(list /ipfs/QmHash1 /ipfs/QmHash2 /ipld/QmHash3)

;; Access the underlying kubo path
(.Path my-path)  ; Returns the kubo path.Path object

;; Convert to string for compatibility
(.String my-path)  ; Returns the path as a string
```

### Implementation Details

The `UnixPathReader` function returns a `reader.Macro` that:

1. Reads characters until it encounters whitespace or delimiters
2. Validates that the path starts with `/ipfs/` or `/ipld/`
3. Uses kubo's `path.NewPath()` to validate the path format and CID encoding
4. Returns a `UnixPath` object containing the validated path

The `UnixPath` type provides:
- `String()` - Returns the path as a string
- `Path()` - Returns the underlying kubo `path.Path` object
- `ToBuiltinString()` - Converts to `builtin.String` for shell compatibility

This provides robust validation and type safety while maintaining seamless integration with existing IPFS operations. 