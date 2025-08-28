# ww run

Execute binaries in a jailed subprocess with file descriptor passing support.

## Usage

```bash
ww run [flags] <binary> [args...]
```

## File Descriptor Passing

Use `--with-fd` to pass file descriptors to the child process:

```bash
# Single FD
ww run --with-fd db=3 ./app

# Multiple FDs
ww run --with-fd db=3 --with-fd cache=4 ./app

# Different ordering (same result)
ww run --with-fd cache=4 --with-fd db=3 ./app
```

### FD Layout

The child process receives file descriptors in this order:

- **FD 3**: RPC socket for host communication (always present)
- **FD 4**: First `--with-fd` mapping
- **FD 5**: Second `--with-fd` mapping
- **FD 6**: Third `--with-fd` mapping
- And so on...

### Environment Variables

The child receives environment variables mapping names to FD numbers:

```bash
WW_FD_DB=4       # "db" available at FD 4
WW_FD_CACHE=5    # "cache" available at FD 5
```

### Child Process Usage

```go
package main

import (
    "fmt"
    "os"

    "github.com/wetware/go/util"
)

func main() {
    // Get all available file descriptors
    fdMap := util.GetFDMap()
    
    // Check for specific FD
    if dbFD, exists := fdMap["db"]; exists {
        dbFile := os.NewFile(uintptr(dbFD), "database")
        defer dbFile.Close()
        // Use dbFile for database operations
        fmt.Printf("Using database at FD %d\n", dbFD)
    } else {
        fmt.Fprintf(os.Stderr, "Database FD not provided\n")
        os.Exit(1)
    }
    
    // Or iterate through all available FDs
    for name, fd := range fdMap {
        fmt.Printf("Available: %s -> FD %d\n", name, fd)
    }
}
```

## Flags

- `--with-fd name=fdnum`: Map parent FD to child with semantic name
- `--env key=value`: Set environment variable for child process
- `--ipfs addr`: IPFS API endpoint (default: `/dns4/localhost/tcp/5001/http`)

## Implementation

The `--with-fd` system:

1. **Parses** `--with-fd` flags into name/fd pairs
2. **Validates** names (no duplicates) and FD numbers (non-negative)
3. **Duplicates** source FDs using `syscall.Dup()` to avoid conflicts
4. **Assigns** target FDs sequentially starting at 4
5. **Generates** environment variables for child process
6. **Passes** FDs via `cmd.ExtraFiles`
7. **Cleans up** resources when done

## Examples

### Basic Usage
```bash
# Run local binary
ww run ./myapp

# Run IPFS binary
ww run /ipfs/QmHash.../myapp

# With environment variables
ww run --env DEBUG=1 --env LOG_LEVEL=info ./myapp
```

### File Descriptor Examples
```bash
# Pass database socket
ww run --with-fd db=3 ./database-app

# Pass multiple resources
ww run --with-fd db=3 --with-fd cache=4 --with-fd logs=5 ./full-app

# Combine with environment
ww run --with-fd db=3 --env DB_TIMEOUT=30s ./db-app
```

## Testing

```bash
# Run tests
go test ./cmd/ww/run/...

# Build
go build ./cmd/ww/...
```

## Demo

See `examples/fd-demo/` for a working example of file descriptor usage.
