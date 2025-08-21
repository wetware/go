# Wetware Go

Go implementation of the Wetware distributed computing system.

## Installation

```bash
go install github.com/wetware/go/cmd/ww
```

## Commands

- `ww run <binary>` - Execute binaries in isolated cells with IPFS support
- `ww export <path>` - Add files/directories to IPFS
- `ww import <ipfs-path>` - Download content from IPFS
- `ww idgen` - Generate Ed25519 private keys

## Architecture

Wetware provides capability-based security through isolated execution environments (cells) with controlled access to IPFS and other distributed services. Each cell runs in a jailed process with file descriptor-based capability passing.

See [spec/cell.md](spec/cell.md) for the cell API specification.
