# Wetware Go

Go implementation of the Wetware distributed computing system.

## Installation

```bash
go install github.com/wetware/go/cmd/ww
```

## Quick Start

Run WASM binaries directly with `ww run`:

```bash
# Run a local WASM file
ww run ./myapp.wasm

# Run from IPFS
ww run /ipfs/QmHash/myapp.wasm

# Run from $PATH
ww run myapp

# Run with debug info
ww run --wasm-debug ./myapp.wasm
```

## Commands

- `ww run <binary>` - Execute WASM binaries with libp2p networking
- `ww shell` - Interactive LISP shell with IPFS and P2P capabilities
- `ww export <path>` - Add files/directories to IPFS
- `ww import <ipfs-path>` - Download content from IPFS
- `ww idgen` - Generate Ed25519 private keys

## Architecture

Wetware provides capability-based security through WASM-based execution environments with controlled access to IPFS and other distributed services. Each WASM module runs with its `poll()` export served on libp2p streams at `/ww/0.1.0/{proc-id}`.

### WASM Process Model

- **Binary Resolution**: Supports local files, $PATH binaries, and IPFS paths
- **WASM Runtime**: Uses wazero for secure WASM execution
- **libp2p Integration**: Serves WASM `poll()` export on network streams
- **IPFS Support**: Direct access to IPFS for distributed content

## Examples

### Hello World WASM

Build and run a simple WASM module:

```bash
# Install tinygo
go install tinygo.org/x/tinygo@latest

# Build example
cd examples/hello
tinygo build -target wasi -o hello.wasm main.go

# Run with ww
ww run hello.wasm
```

See [examples/hello/README.md](examples/hello/README.md) for more details.

## Documentation

- [Cell API Specification](spec/cell.md) - Complete API reference
- [Shell Documentation](cmd/ww/shell/README.md) - Interactive shell guide
