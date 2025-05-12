# Glia

Glia is an actor concurrency protocol implemented over libp2p, where processes are WebAssembly modules running in a WASI environment.

## Overview

Glia provides a message-passing protocol for distributed actor systems, where:
- Actors are WASM/WASI modules
- Messages are delivered to WASI exported functions
- Communication happens over libp2p
- Both P2P and HTTP interfaces are supported

## Architecture

### Core Concepts

1. **Processes (Actors)**
   - WebAssembly modules running in WASI environments
   - Each process is an independent actor
   - Processes are identified by their Process ID (PID)

2. **Methods**
   - WASI exported functions that handle messages
   - Each method is a behavior that can be invoked on a process
   - Methods are identified by their name

3. **Streams**
   - Represent message channels between processes
   - Carry metadata about the message:
     - Protocol ID
     - Destination
     - Process ID
     - Method name
   - Implement `io.ReadWriteCloser` for message data

### Protocol Flow

1. **Message Delivery**
   ```
   Message -> Stream -> WASI Exported Function
   ```
   - Messages are sent over streams
   - Streams route messages to the correct process and method
   - Methods are invoked as WASI exported functions

2. **Transport Layer**
   - Primary: libp2p for P2P communication
   - Secondary: HTTP API for traditional access
   - Both interfaces use the same underlying protocol

## Implementation

### Environment

The `Env` interface provides core services:
```go
type Env interface {
    Log() *slog.Logger
    LocalHost() host.Host
    Routing() core_routing.Routing
}
```

### Stream Interface

The `Stream` interface defines message channels:
```go
type Stream interface {
    Protocol() protocol.ID
    Destination() string
    ProcID() string
    MethodName() string
    io.ReadWriteCloser
    CloseRead() error
    CloseWrite() error
}
```

### HTTP API

The HTTP interface provides REST endpoints:
- `/status`: API path with root PID
- `/info`: Host peer information
- `/version`: Version information
- `/root`: Root information
- Dynamic endpoint for glia operations

## Usage

### Creating a Process

1. Create a WASM module with exported functions
2. Register the process with the Glia runtime
3. The exported functions become available as methods

### Sending Messages

1. Create a stream to the target process
2. Specify the method name
3. Send the message data
4. Handle the response

## Security

- Processes are sandboxed by WASI
- Communication is secured by libp2p
- Access control can be implemented at the process level

## Testing

The package includes comprehensive tests:
- Unit tests for core functionality
- Integration tests for HTTP and P2P interfaces
- Mock implementations for testing

## License

[License information]
