# Wetware Run Command

The `ww run` command executes binaries from both local filesystem and IPFS paths, providing seamless integration between traditional file systems and distributed storage.

## Overview

The run command supports three types of executable sources:

1. **Local filesystem paths** (existing behavior)
2. **IPFS file paths** (single executable files)
3. **IPFS directory paths** (with Go-style OS/architecture organization)

## Basic Usage

```bash
# Run a local binary
ww run ./my-program args...

# Run from IPFS file path
ww run /ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa args...

# Run from IPFS directory
ww run /ipfs/QmSomeDirectoryHash/ args...
```

## IPFS Path Support

### File Paths

When you provide an IPFS path pointing to a file, `ww run` will:

1. Download the file to a temporary directory
2. Make it executable (`chmod +x`)
3. Execute it with the provided arguments

```bash
# Example: Running a single binary from IPFS
ww run /ipfs/QmAbC123.../hello-world --greeting="Hello, IPFS!"
```

### Directory Paths

For IPFS directories, `ww run` follows Go's OS/architecture convention by looking for binaries in a `bin/` subdirectory organized by operating system and architecture.

Expected directory structure:
```
/ipfs/QmDirectoryHash/
├── bin/
│   ├── darwin_amd64/
│   │   └── my-program
│   ├── darwin_arm64/
│   │   └── my-program
│   ├── linux_amd64/
│   │   └── my-program
│   ├── linux_arm64/
│   │   └── my-program
│   └── windows_amd64/
│       └── my-program.exe
└── README.md
```

The command will:

1. Extract the directory structure to a temporary location
2. Look for the executable in `bin/{OS}_{ARCH}/`
3. Fall back to looking directly in `bin/` if OS/arch-specific version not found
4. Make the binary executable and run it

```bash
# Example: Running from a directory with multiple OS/arch binaries
ww run /ipfs/QmDirectoryHash/ --config=production
```

## Environment Variables

You can override the default OS and architecture detection:

- **`WW_OS`**: Override the operating system (defaults to `runtime.GOOS`)
- **`WW_ARCH`**: Override the architecture (defaults to `runtime.GOARCH`)

```bash
# Force Linux AMD64 binary even on different platform
WW_OS=linux WW_ARCH=amd64 ww run /ipfs/QmDirectoryHash/

# Use ARM64 binary
WW_ARCH=arm64 ww run /ipfs/QmDirectoryHash/
```

## File Descriptor Passing

The `ww run` command supports secure file descriptor passing, allowing you to grant specific file descriptors to child processes with fine-grained control over access modes and types.

### Basic Usage

```bash
# Pass a database socket
ww run --fd db=3,mode=rw,type=socket /ipfs/foo

# Pass multiple file descriptors
ww run \
  --fd db=3,mode=rw,type=socket \
  --fd logs=5,mode=w,type=file \
  /ipfs/foo
```

### Advanced Features

- **Bulk configuration** via S-expression files
- **Systemd socket activation** support
- **Automatic target fd assignment** (starts from 10)
- **Symlink creation** in jail directories
- **Verbose logging** of fd grants

For complete documentation, see the [main README](../../../README.md#file-descriptor-passing).

## Fallback Behavior

The command uses intelligent fallback logic:

1. **IPFS Path Parsing**: First attempts to parse the path as an IPFS path
2. **Local Filesystem**: If IPFS parsing fails, treats it as a local filesystem path
3. **Path Resolution**: Relative local paths are resolved to absolute paths
4. **Error Reporting**: Clear error messages if neither IPFS nor local path works

```bash
# This will try IPFS first, then fall back to local filesystem
ww run my-local-binary

# This will be treated as a local path (not valid IPFS format)
ww run ./scripts/deploy.sh
```

## Error Handling

Common error scenarios and their meanings:

- **"failed to get IPFS path"**: The IPFS path exists but couldn't be retrieved (network issues, missing content)
- **"no executable found in IPFS directory"**: Directory doesn't contain expected `bin/` structure
- **"unexpected node type"**: IPFS path points to unsupported content type
- **"failed to resolve path"**: Local filesystem path issues

## Examples

### Publishing a Binary to IPFS

```bash
# Add your binary to IPFS
ipfs add my-program
# Returns: QmAbC123...

# Run it directly
ww run /ipfs/QmAbC123... --help
```

### Creating a Multi-Platform Distribution

```bash
# Create directory structure
mkdir -p dist/bin/{darwin_amd64,darwin_arm64,linux_amd64,linux_arm64,windows_amd64}

# Build for different platforms
GOOS=darwin GOARCH=amd64 go build -o dist/bin/darwin_amd64/my-app
GOOS=darwin GOARCH=arm64 go build -o dist/bin/darwin_arm64/my-app
GOOS=linux GOARCH=amd64 go build -o dist/bin/linux_amd64/my-app
GOOS=linux GOARCH=arm64 go build -o dist/bin/linux_arm64/my-app
GOOS=windows GOARCH=amd64 go build -o dist/bin/windows_amd64/my-app.exe

# Add to IPFS
ipfs add -r dist/
# Returns: QmDistHash...

# Run platform-appropriate binary
ww run /ipfs/QmDistHash/ --version
```

### Development Workflow

```bash
# During development, use local paths
ww run ./bin/my-app --debug

# For testing distribution
ipfs add ./bin/my-app
ww run /ipfs/QmNewHash... --debug

# For production deployment
ww run /ipfs/QmProductionHash/ --config=prod
```

## Integration with Wetware

The run command integrates seamlessly with Wetware's capabilities:

- **Cell Isolation**: Each executed binary runs in its own isolated cell
- **Capability Security**: Binaries receive only the capabilities they need
- **P2P Communication**: Executed programs can communicate over libp2p protocols
- **Resource Management**: Automatic cleanup of temporary files and directories

## Configuration

The command supports standard Wetware flags:

```bash
# Custom IPFS endpoint
ww run --ipfs=/ip4/127.0.0.1/tcp/5001/http /ipfs/QmHash...

# Environment variables for the executed program
ww run --env=MY_VAR=value --env=DEBUG=true /ipfs/QmHash...
```

## Security Considerations

- **Code Verification**: Always verify the integrity of IPFS content before execution
- **Network Trust**: IPFS content is retrieved from the network; ensure your IPFS node connects to trusted peers
- **Execution Environment**: Binaries run with the same privileges as the `ww` process
- **Temporary Files**: Downloaded content is automatically cleaned up after execution

---

For more information about Wetware and its capabilities, see the main project documentation.
