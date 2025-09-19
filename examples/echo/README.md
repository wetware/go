# Echo Example

This example demonstrates the idiomatic wetware process pattern for supporting both synchronous and asynchronous execution modes.

## Pattern Overview

Wetware processes follow a unified pattern that supports both sync and async behaviors through a single `Proc` configuration:

1. **Sync Mode** (`Async: false`): The `_start` function runs automatically, calling `main()` which processes stdin and exits
2. **Async Mode** (`Async: true`): The `_start` function is prevented from running, and the `poll()` export is called for each incoming stream

## Implementation

### Sync Mode
```go
func main() {
    // Process stdin and exit
    if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
        os.Stderr.WriteString("Error: " + err.Error() + "\n")
        os.Exit(1)
    }
    // Return 0 to indicate successful completion
}
```

### Async Mode
```go
//export poll
func poll() {
    // Process each incoming stream
    if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
        os.Stderr.WriteString("Error in poll: " + err.Error() + "\n")
        os.Exit(1)
    }
}
```

## Runtime Behavior

### Sync Mode (`Async: false`)
1. **Module Instantiation**: `_start` runs automatically during `ProcConfig.New()`
2. **Message Processing**: `main()` reads from stdin until EOF (one complete message)
3. **Module Closure**: Module closes after `main()` returns
4. **Usage**: One module instance per message

### Async Mode (`Async: true`)
1. **Module Instantiation**: `_start` is prevented from running (`WithStartFunctions()`)
2. **Stream Handler**: Module is registered with a stream handler
3. **Message Processing**: Each incoming stream calls `poll()` export
4. **Module Persistence**: Module stays alive for multiple messages
5. **Usage**: One module instance for multiple messages

## Configuration

```go
// Sync mode
config := system.ProcConfig{
    Host:      host,
    Runtime:   runtime,
    Bytecode:  bytecode,
    ErrWriter: &bytes.Buffer{},
    Async:     false, // Sync mode
}

// Async mode
config := system.ProcConfig{
    Host:      host,
    Runtime:   runtime,
    Bytecode:  bytecode,
    ErrWriter: &bytes.Buffer{},
    Async:     true, // Async mode
}
```

## Benefits

- **Explicit Mode Selection**: `Async` flag makes behavior clear and predictable
- **WASM Lifecycle Compliance**: Respects the fundamental constraint that `_start` can only run once
- **Flexible Implementation**: Processes can support one or both modes
- **Clean Separation**: Sync and async logic are clearly separated
- **Unified API**: Same `ProcessMessage` method works for both modes

## Usage

### Building
```bash
tinygo build -o main.wasm -target=wasi -scheduler=none main.go
```

### Running
The process behavior is determined by the `Async` configuration flag:
- **Sync mode**: Process one message and exit
- **Async mode**: Process multiple messages via stream handler

This pattern enables wetware processes to be both simple command-line tools and long-running stream processors, depending on the configuration and requirements.