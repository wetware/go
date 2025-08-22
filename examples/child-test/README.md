# Child Cell Example

This example demonstrates the `--with-child` functionality of `ww run` for composing multiple cells.

## Building

```bash
go build -o child-test
```

## Usage

```bash
# Run with two child cells
ww run --with-child "db=./child-test" --with-child "cache=./child-test" ./child-test

# Run with IPFS-based children
ww run --with-child "db=/ipfs/QmHash.../db-service" --with-child "cache=/ipfs/QmHash.../cache-service" ./child-test

# Run with file descriptor-based children
ww run --with-child "storage=3" --with-child "network=4" ./child-test
```

## What Happens

1. **Main process starts** - The specified binary runs in the primary cell
2. **Child cells start** - Each `--with-child` flag creates a child cell
3. **Environment variables** - Each child receives:
   - `WW_CHILD_NAME` - The name specified (e.g., "db", "cache")
   - `WW_CHILD_INDEX` - Sequential index starting from 0
4. **Capability sharing** - Children can communicate with parent via socket pairs

## Child Cell Types

- **IPFS paths**: `/ipfs/QmHash.../service` - Downloads and executes from IPFS
- **Local paths**: `./local/service` - Executes from local filesystem  
- **File descriptors**: `3` - Uses existing open file descriptor (must be > 2)

## Example Output

```
Child cell started!
Child name: 
Child index: 
Parent PID: 12345
Child cell running... (press Ctrl+C to stop)

Child cell started!
Child name: db
Child index: 0
Parent PID: 12345
Child cell running... (press Ctrl+C to stop)

Child cell started!
Child name: cache
Child index: 1
Parent PID: 12345
Child cell running... (press Ctrl+C to stop)
```
