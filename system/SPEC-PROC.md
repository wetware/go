# Wetware Process Specification

This document specifies the design and behavior of Wetware processes (`Proc`) in the wetware system.

## Overview

A Wetware process is a WebAssembly (WASM) module that can operate in two distinct modes:
- **Synchronous Mode**: Processes one message and exits
- **Asynchronous Mode**: Processes multiple messages via stream handlers

The mode is determined by the `Async` field in `ProcConfig`.

## Core Design Principles

### 1. WASM Lifecycle Constraint
WebAssembly modules have a fundamental constraint: `_start` can only run once during module instantiation. After `_start` returns, the module is closed and cannot be reused.

This constraint forces us to choose between **at runtime**:
- **Running `_start`** (sync mode): Module processes one message and closes
- **Preventing `_start`** (async mode): Module stays alive, exports are called repeatedly

Note the bold lettering above:  a single WASM executable can support **both** synchronous and asynchronous modes.  Usually, this maps onto a **client and server mode**.

>**Recommendation.**  Structure your Wetware applications as single-binary executables that behave as a server in async mode, and as a client in sync mode.

### 2. Message Processing Protocol
**One stream (start to EOF) is one message.** This is the fundamental protocol for message delivery:

- A complete message is defined as data read from a network stream from start until EOF
- The WASM module reads from stdin until EOF to consume one complete message
- After EOF, the message processing is complete
- In async mode, the next call to `poll()` will receive a new stream with a new message

## Synchronous Mode (`Async: false`)

### Behavior
- `_start` runs automatically during module instantiation (i.e. `main()` runs exactly once)
- `main()` function processes one complete message from stdin (start to EOF)
- Module closes after `main()` returns
- One module instance per message

### Message Delivery Mechanism
1. **Stream Setup**: A network stream is connected to stdin before module instantiation
2. **Message Processing**: `main()` reads from stdin until EOF (one complete message)
3. **Module Closure**: Module closes after `main()` returns
4. **Next Message**: Requires a new module instance

### Guest Code Pattern
The guest code implements a simple message processor:
```go
func main() {
    // Process one complete message from stdin (start to EOF)
    if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
        os.Stderr.WriteString("Error: " + err.Error() + "\n")
        os.Exit(1)
    }
    // Return 0 to indicate successful completion
}
```

## Asynchronous Mode (`Async: true`)

### Behavior
- `_start` is prevented from running during module instantiation (i.e. `main()` will not run at all)
- Module stays alive for multiple messages
- Each incoming stream calls the `poll()` export function
- One module instance for multiple messages

### Message Delivery Mechanism
1. **Module Instantiation**: Module is created without running `_start` (main() never runs)
2. **Stream Handler Registration**: Module is registered to handle incoming streams
3. **Message Processing**: Each new stream triggers a call to `poll()`
4. **Stream Consumption**: `poll()` reads from stdin until EOF (one complete message)
5. **Stream Completion**: After EOF, the stream is closed and `poll()` returns
6. **Next Message**: A new stream triggers another `poll()` call

### Guest Code Pattern
The guest code implements a stream-based message processor:
```go
//export poll
func poll() {
    // Process one complete message from stdin (start to EOF)
    if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
        os.Stderr.WriteString("Error in poll: " + err.Error() + "\n")
        os.Exit(1)
    }
}
```

## Message Delivery Protocol

### Stream-to-Message Mapping
- **One Network Stream = One Message**
- **Stream Start**: Beginning of message data
- **Stream EOF**: End of message data
- **Message Boundary**: EOF marks the complete message

### Stream Handling
1. **Stream Connection**: Network stream is connected to stdin
2. **Message Reading**: WASM module reads from stdin until EOF
3. **Message Processing**: Complete message is processed by the module
4. **Stream Cleanup**: Stream is closed after EOF
5. **Next Message**: New stream triggers next message processing

### Error Handling
- **Stream Errors**: Network errors during reading are propagated to the module
- **Processing Errors**: Module errors are logged and may cause module termination
- **Timeout Handling**: Stream timeouts are handled according to context deadlines

## Configuration

### ProcConfig
```go
type ProcConfig struct {
    Host      host.Host
    Runtime   wazero.Runtime
    Bytecode  []byte
    ErrWriter io.Writer
    Async     bool // Gates sync vs async behavior
}
```

### Mode Selection
- **Sync Mode** (`Async: false`): One message per module instance
- **Async Mode** (`Async: true`): Multiple messages per module instance

## Benefits

1. **Explicit Mode Selection**: `Async` flag makes behavior clear and predictable
2. **WASM Lifecycle Compliance**: Respects fundamental WASM constraints
3. **Simple Message Protocol**: One stream to EOF is one message
4. **Flexible Implementation**: Processes can support one or both modes
5. **Clean Separation**: Sync and async logic are clearly separated
6. **Unified API**: Same `ProcessMessage` method works for both modes

## Use Cases

### Synchronous Mode
- **Command-line tools**: Process one input and exit
- **Batch processing**: Process one file or data stream
- **Simple transformations**: One-shot data processing

### Asynchronous Mode
- **Stream processors**: Handle multiple incoming streams
- **Long-running services**: Persistent message processing
- **Real-time systems**: Continuous message handling

## Conclusion

The dual-mode design provides a clean, predictable way to handle both synchronous and asynchronous WebAssembly processes while respecting the fundamental constraints of the WASM runtime. The explicit `Async` flag makes the behavior clear, and the simple "one stream to EOF is one message" protocol provides a consistent interface for message delivery across both modes.