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

## Go Special Form

The `go` special form allows execution of code in a different process context, enabling secure, capability-based cross-process communication. It supports both local and remote executors and can be used for parallel processing across multiple hosts.

### Usage

```lisp
;; Basic usage - spawns executable in a separate cell on the local host
(go /ipfs/QmWKKmjmTmbaFuU4Bu92KXob3jaqKJ9vZXRch6Ks8GJESZ/cmd/shell
    (console.Println "Hello, World!")
    :console console)

;; Using a remote peer executor
(go /ipfs/QmWKKmjmTmbaFuU4Bu92KXob3jaqKJ9vZXRch6Ks8GJESZ/cmd/shell
    (console.Println "Hello, World!")
    :exec peer-executor
    :console console)

;; More complex example with data processing
(go /ipfs/QmWKKmjmTmbaFuU4Bu92KXob3jaqKJ9vZXRch6Ks8GJESZ/cmd/data-processor
    (-> (source.FetchData "sensor-readings")
        (map #(assoc % :processed true))
        (results.Send))
    :exec cluster-executor
    :data-source source
    :result-channel results)

;; Concurrent processing example
(let [results (chan)]
  (go /ipfs/QmWKKmjmTmbaFuU4Bu92KXob3jaqKJ9vZXRch6Ks8GJESZ/cmd/worker
    (process-work work-item)
    :exec worker-pool
    :work-item task1
    :result-channel results)
  
  (go /ipfs/QmWKKmjmTmbaFuU4Bu92KXob3jaqKJ9vZXRch6Ks8GJESZ/cmd/worker
    (process-work work-item)
    :exec worker-pool
    :work-item task2
    :result-channel results))
```

### Parameters

- **executable path** (first argument): Must be a valid IPFS path (`UnixPath`) pointing to the executable to spawn
- **body** (second argument): The code to execute in the spawned context
- **keyword arguments** (optional):
  - `:exec` - Executor capability for the target host (defaults to local executor)
  - `:console` - Console capability for output
  - Additional capabilities can be passed as keyword arguments

### Key Features

- **Cross-process execution**: Allows execution of code in a different process context
- **Capability-based security**: Supports passing capabilities via keyword arguments
- **Local and remote execution**: Works with both local and remote executors
- **Parallel processing**: Enables concurrent processing across multiple hosts
- **Secure communication**: Uses capability-based cross-process communication

### Implementation Details

The `go` special form:

1. Validates the executable path as a `UnixPath`
2. Parses keyword arguments for capabilities and additional arguments
3. Serializes the body to a string representation
4. Spawns a new process using the executor capability with the serialized body as an argument
5. Returns a `Cell` object representing the spawned process

The spawned process runs in a controlled execution environment with attenuated capabilities, providing security and isolation while enabling powerful distributed computing capabilities.

### Current Limitations

- The body is serialized as a string and passed as a command-line argument
- Capability passing is not yet fully implemented
- The spawned process must be able to parse the serialized body format 