# Wetware Cell API Specification

## Motivation

The Wetware Cell API defines the interface between a host process (running `ww run`) and a cell process (child process managed by `ww run`) that runs in a controlled execution environment with attenuated capabilities.

This document is for users who want to write their own cells. Simply adhere to this specification, and then you can `ww run <yourbinary>`, and your application will run in a secure, decentralized environment with access to capabilities like IPFS.

## Process Interface

### Standard Input/Output
- **stdin (fd 0)**: Direct passthrough from host's stdin
- **stdout (fd 1)**: Direct passthrough to host's stdout  
- **stderr (fd 2)**: Direct passthrough to host's stderr

### Command Line Arguments
- **argv[0]**: Executable path
- **argv[1..n]**: Arguments passed from host to cell
- Arguments are passed through unchanged from host's `exec.Command`

### Environment Variables
- **WW_ENV**: Comma-separated list of environment variables to pass to cell
- Additional environment variables can be specified via `--env` flag
- Cell inherits host's environment filtered by these specifications

## File Descriptors

### Fixed File Descriptors
- **fd 3**: One end of a Unix domain socket pair for Cap'n Proto RPC communication with host

### User-Configurable File Descriptors
Additional file descriptors can be passed from host to cell using the `--fd` flag:

```bash
ww run --fd name=fdnum [--fd name2=fdnum2 ...] <command>
```

**Format**: `--fd name=fdnum` where:
- `name`: Logical name for the file descriptor (e.g., "db", "cache", "input")
- `fdnum`: Source file descriptor number in the host process

**Target Assignment**: File descriptors are automatically assigned to the cell in **predictable positional order** starting at fd 3:
- First `--fd` flag → fd 3 (first argument)
- Second `--fd` flag → fd 4 (second argument)  
- Third `--fd` flag → fd 5 (third argument)
- And so on...

**Naming Benefits**: Each file descriptor gets a semantic name that the cell can use:
- **Predictable positioning**: The child knows exactly which FD number to use
- **Semantic meaning**: Names like "db", "cache", "logs" make the purpose clear
- **Flexible ordering**: You can reorder `--fd` flags without breaking the child
- **Environment mapping**: Clear mapping from names to FD numbers

**Environment Variables**: The cell receives environment variables that map names to FD numbers:
```bash
WW_FD_DB=3       # "db" is available at FD 3
WW_FD_CACHE=4    # "cache" is available at FD 4
WW_FD_INPUT=5    # "input" is available at FD 5
```

**Usage in Child Process**: The cell can now access resources by both name and position:
```go
// Access by FD number (predictable positioning)
dbFD := os.NewFile(3, "database")
cacheFD := os.NewFile(4, "cache")

// Or check environment variables for validation
if os.Getenv("WW_FD_DATABASE") == "" {
    log.Fatal("Database FD not provided")
}
```

### File Descriptor Usage
The cell process must:
1. Establish RPC connection via fd 3
2. Authenticate to obtain capabilities
3. Access additional file descriptors via predictable FD numbers (3, 4, 5...) or environment variables (`WW_FD_*`)

**Key Benefits**:
- **Predictable access**: First `--fd` is always at FD 3, second at FD 4, etc.
- **Named resources**: Environment variables tell you what each FD represents
- **Flexible ordering**: Change the order of `--fd` flags without breaking the child
- **Self-documenting**: The child knows exactly what each FD is for

## Capabilities

### Available Capabilities
Currently, exactly one capability is available:

- **IPFS**: Access to IPFS Core API
  - Interface: `system.IPFS` (Cap'n Proto)
  - Access: Policy-controlled based on authentication
  - Scope: Determined by host's IPFS configuration

### Future Capabilities
The interface is designed to support additional capabilities:
- Process execution
- Network access (via libp2p streams)
- Various decentralized services

## Authentication

*Note: Authentication details are currently being simplified. The specific authentication mechanism may vary between implementations.*

### Current Implementation
The current `ww run` implementation uses a simplified approach where:
- The cell connects to the host via the Unix domain socket (fd 3)
- Capabilities are granted based on the connection establishment
- No separate identity file or cryptographic authentication is required

### Future Authentication
Future versions may implement more sophisticated authentication mechanisms including:
- Cryptographic identity verification
- Policy-based capability grants
- Multi-factor authentication

## Status Codes

### Standard Exit Codes
- **0**: Success
- **1**: General error
- **2**: Usage error
- **126**: Command not executable
- **127**: Command not found
- **128+n**: Signal termination (n = signal number)

### Future Standard Codes
The following status codes are reserved for future standardization:
- **64**: Cell authentication failed
- **65**: Capability access denied
- **66**: Resource limit exceeded
- **67**: Isolation violation detected

## Reference Implementations

### ww run
The `ww run` subcommand is a reference implementation that demonstrates the cell API. It:
- Creates a jailed execution environment
- Sets up the Unix domain socket pair for RPC communication
- Launches the specified executable with the required file descriptors
- Supports file descriptor passing via `--fd` flags

**Example Usage**:
```bash
# Example usage:
# ww run --fd db=3 --fd cache=4 /ipfs/QmMyApp
# 
# Environment variables in cell:
# - WW_FD_DB=3 (fd 3 mapped to fd 3 in cell)
# - WW_FD_CACHE=4 (fd 4 mapped to fd 4 in cell)
```
