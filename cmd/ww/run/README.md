# Run Command

Executes binaries in isolated cells with IPFS support and capability-based security.

## Usage

```bash
ww run <binary> [args...]
```

## Arguments

- `<binary>` - Executable path (local filesystem or IPFS path)
- `[args...]` - Arguments passed to the binary

## Options

- `--ipfs` - IPFS API endpoint (default: `/dns4/localhost/tcp/5001/http`)
- `--env` - Environment variables for the cell (can be specified multiple times)
- `--with-fd <name>=<fdnum>` - Map parent file descriptor to name (can be specified multiple times)
- `WW_IPFS` environment variable overrides `--ipfs` flag
- `WW_ENV` environment variable provides comma-separated list of environment variables

## IPFS Path Support

When `<binary>` is an IPFS path pointing to a file:

1. Downloads the file to a temporary directory
2. Makes it executable (`chmod +x`)
3. Executes it with provided arguments

```bash
# Run binary from IPFS
ww run /ipfs/QmHash.../my-program --help
```

## File Descriptor Management

### Basic Syntax

```bash
ww run --with-fd <name>=<fdnum> <binary>
```

- **`<name>`** - Logical name for the file descriptor
- **`<fdnum>`** - Source file descriptor number in parent process

### Target Assignment

File descriptors are automatically assigned to the cell in predictable order starting at fd 3:

- First `--with-fd` flag → fd 3
- Second `--with-fd` flag → fd 4
- Third `--with-fd` flag → fd 5
- And so on...

**Note**: FDs are assigned in the order they appear in the child's `ExtraFiles` slice, ensuring consistent numbering regardless of command-line flag order.

### Environment Variables

The cell receives environment variables mapping names to FD numbers:

```bash
WW_FD_DB=3       # "db" is available at FD 3
WW_FD_CACHE=4    # "cache" is available at FD 4
WW_FD_INPUT=5    # "input" is available at FD 5
```

### Examples

```bash
# Pass database socket and cache
ww run --with-fd db=3 --with-fd cache=4 /ipfs/foo

# Pass multiple file descriptors
ww run \
  --with-fd db=3 --with-fd logs=5 \
  /ipfs/foo
```

## Cell Environment

### Standard File Descriptors

- **stdin (fd 0)**: Direct passthrough from host
- **stdout (fd 1)**: Direct passthrough to host
- **stderr (fd 2)**: Direct passthrough to host
- **fd 3**: Unix domain socket for Cap'n Proto RPC with host

### Capabilities

Cells connect to the host via fd 3 to obtain capabilities:

- **IPFS**: Access to IPFS Core API via `system.IPFS` interface
- **Future**: Process execution, network access, decentralized services

### Authentication

Current implementation uses simplified authentication:
- Cell connects to host via Unix domain socket (fd 3)
- Capabilities granted based on connection establishment
- No separate identity file or cryptographic authentication required

## Implementation Details

1. **Path Resolution**: Attempts IPFS path parsing first, falls back to local filesystem
2. **Cell Creation**: Sets up jailed execution environment with Unix domain socket pair
3. **File Descriptor Preparation**: Maps specified FDs to predictable positions
4. **Process Execution**: Launches binary with configured environment and FDs
5. **RPC Setup**: Establishes libp2p protocol handler for `/ww/0.1.0`

## Error Handling

- IPFS retrieval errors: "failed to get IPFS path"
- Unexpected content types: "unexpected node type"
- Path resolution errors: "failed to resolve path"
- FD processing errors: "fd flag processing failed"

## Examples

### Local Binary

```bash
# Run local binary
ww run ./bin/my-app --debug

# Run with environment variables
ww run --env DEBUG=true --env LOG_LEVEL=debug ./my-app
```

### IPFS Binary

```bash
# Export and run binary
ww export ./my-program
# Output: /ipfs/QmHash...

ww run /ipfs/QmHash... --help
```

### File Descriptor Passing

```bash
# Pass database socket to cell
ww run --with-fd db=3 /ipfs/QmHash.../database-app

# Cell can access database at fd 3
# Environment: WW_FD_DB=3
```

## Requirements

- IPFS daemon running and accessible
- Proper IPFS API endpoint configuration
- Sufficient permissions for jailed execution
- Target binary must be executable

## Security Considerations

- Cells run in isolated environments with attenuated capabilities
- File descriptors provide controlled access to parent resources
- IPFS content integrity should be verified before execution
- No automatic content verification or sandboxing beyond basic isolation
