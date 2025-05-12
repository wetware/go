# Glia Protocol Specification

## Overview

Glia is an actor concurrency protocol implemented over libp2p, where processes are WebAssembly modules running in a WASI environment.

## Core Concepts

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

## Protocol Flow

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

## Frame Encoding

The Glia protocol uses a simple frame encoding for messages:

- Each frame consists of a uvarint length prefix followed by the message data.
- The uvarint is encoded using the standard Go `binary.PutUvarint` function.
- The message data is written directly after the uvarint.

### Example

For a message of length 5 bytes:

1. Encode the length as a uvarint: `[5]` (1 byte)
2. Write the message data: `[data]` (5 bytes)
3. Total frame: `[5][data]` (6 bytes)

### Reading Frames

To read a frame:

1. Read the uvarint length prefix.
2. Allocate a buffer of the specified length.
3. Read the message data into the buffer.

### Error Handling

- If the uvarint is truncated, an error is returned.
- If the message data is truncated, an error is returned.
- If the message size is unreasonably large, an error is returned.

### Concurrency

The frame encoding does not guarantee safe concurrent access. Ensure proper synchronization if multiple goroutines are writing to the same buffer.

## Security

- Processes are sandboxed by WASI
- Communication is secured by libp2p
- Access control can be implemented at the process level

## License

[License information] 