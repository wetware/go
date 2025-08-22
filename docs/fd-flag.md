# File Descriptor Flag (`--with-fd`) for `ww run`

This document describes the `--with-fd` flag implementation in the `ww run` command, which provides a clean, predictable way to grant child processes access to specific file descriptors with semantic naming.

## Usage

```bash
ww run --with-fd <name>=<fdnum> <binary>
```

## Behavior

- **FD Assignment**: FDs are assigned sequentially starting at 3 (first → fd 3, second → fd 4, etc.)
- **Environment Variables**: Child receives `WW_FD_<NAME>=<target_fd>` for each named FD
- **Validation**: Prevents duplicate names, validates FD numbers
- **Order**: FDs are assigned in the order they appear in the child's `ExtraFiles` slice, ensuring consistent numbering regardless of command-line flag order

## Examples

```bash
# Pass database socket
ww run --with-fd db=3 /ipfs/foo

# Pass multiple FDs
ww run --with-fd db=3 --with-fd cache=4 /ipfs/foo

# Pass multiple FDs with different ordering (same result)
ww run --with-fd cache=4 --with-fd db=3 /ipfs/foo
```

## Implementation Details

### FD Processing Flow

1. **Flag Parsing**: Each `--with-fd` flag is parsed in `name=fdnum` format
2. **Validation**: Checks for duplicate names and valid FD numbers
3. **File Preparation**: Source FDs are duplicated using `syscall.Dup()`
4. **Target Assignment**: FDs are assigned sequentially starting at 3
5. **Environment Generation**: `WW_FD_<NAME>=<target_fd>` variables are created

### Key Implementation Features

- **Deterministic Assignment**: FDs are always assigned in the same order regardless of flag order
- **Safe Duplication**: Uses `syscall.Dup()` to avoid conflicts with parent process
- **Clean Cleanup**: All managed FDs are properly closed when the process exits
- **Error Handling**: Comprehensive validation and error reporting

### Code Structure

- **`FDManager`**: Main struct managing FD configurations and operations
- **`ParseFDFlag()`**: Parses individual `--with-fd` flag values
- **`PrepareFDs()`**: Prepares and assigns target FDs
- **`GenerateEnvVars()`**: Creates environment variables for child process
- **`Close()`**: Cleanup and resource management

## Environment Variables

The child process receives environment variables in this format:

```bash
WW_FD_DB=3       # "db" is available at FD 3
WW_FD_CACHE=4    # "cache" is available at FD 4
WW_FD_INPUT=5    # "input" is available at FD 5
```

## Security Considerations

- **FD Isolation**: Child processes receive duplicated FDs, not direct access to parent FDs
- **Validation**: Prevents duplicate names and validates FD numbers
- **Resource Management**: Proper cleanup ensures no FD leaks

## Testing

The implementation includes comprehensive tests covering:

- Flag parsing and validation
- FD preparation and assignment
- Environment variable generation
- Resource cleanup
- Error handling scenarios

## Compatibility

- **Preserved Functionality**: All existing `ww run` functionality remains intact
- **Backward Compatible**: No breaking changes to existing APIs
- **Enhanced Capability**: Adds new FD passing functionality without affecting other features
