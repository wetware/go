# File Descriptor Demo

This example demonstrates how to use the `--with-fd` flag with `ww run` to pass file descriptors to child processes.

## Building

```bash
go build -o fd-demo
```

## Usage

### Basic Demo (no file descriptors)
```bash
ww run ./fd-demo
```

### With File Descriptors
```bash
# Pass a single file descriptor
ww run --with-fd demo=3 ./fd-demo

# Pass multiple file descriptors
ww run --with-fd demo=3 --with-fd log=4 ./fd-demo

# Pass file descriptors in different order (same result)
ww run --with-fd log=4 --with-fd demo=3 ./fd-demo
```

## What It Does

The demo program:

1. **Checks for RPC socket**: Always available at FD 3 for communication with the host
2. **Lists user FDs**: Shows all file descriptors passed via `--with-fd`
3. **Displays file info**: Attempts to stat each file descriptor to show basic information
4. **Provides guidance**: Shows helpful usage examples if no FDs are provided

## Expected Output

With no file descriptors:
```
File Descriptor Demo
====================
✓ RPC socket available at FD 3

No user file descriptors provided.
Try running with: ww run --with-fd demo=3 --with-fd log=4 ./fd-demo

Demo completed.
```

With file descriptors:
```
File Descriptor Demo
====================
✓ RPC socket available at FD 3
✓ WW_FD_DEMO=4
  └─ File: demo, Size: 1234 bytes
✓ WW_FD_LOG=5
  └─ File: log, Size: 5678 bytes

Demo completed.
```

## File Descriptor Layout

- **FD 3**: RPC socket (always present)
- **FD 4**: First `--with-fd` mapping
- **FD 5**: Second `--with-fd` mapping
- And so on...

## Testing with Real Files

To test with actual files:

```bash
# Create test files
echo "Hello World" > demo.txt
echo "Log message" > log.txt

# Run with file descriptors
ww run --with-fd demo=3 --with-fd log=4 ./fd-demo < demo.txt 3<demo.txt 4<log.txt
```

This demonstrates how the `--with-fd` system works in practice.
