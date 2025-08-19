# Wetware Export Command

The `ww export` command adds files and directories to IPFS recursively, equivalent to `ipfs add -r <path>`.

## Overview

The export command provides a convenient way to publish local files and directories to IPFS, making them available via IPFS paths that can be shared, referenced, or used by other Wetware commands.

## Usage

```bash
ww export <path>
```

## Arguments

- **`<path>`**: The local filesystem path to export. Can be:
  - A file path (e.g., `./document.txt`)
  - A directory path (e.g., `./my-project/`)
  - Current directory (`.`)
  - Absolute paths (e.g., `/home/user/documents`)

## Options

- **`--ipfs`**: IPFS API endpoint (default: `/dns4/localhost/tcp/5001/http`)
  - Can also be set via `WW_IPFS` environment variable
  - Supports multiaddr, URL, or local path formats

## Examples

### Export a single file
```bash
# Export a text file
ww export document.txt
# Output: /ipfs/QmHash...

# Export with custom IPFS endpoint
ww export --ipfs=/ip4/127.0.0.1/tcp/5001/http document.txt
```

### Export a directory
```bash
# Export current directory
ww export .

# Export a specific directory
ww export /path/to/my-project

# Export with environment variable
WW_IPFS=/dns4/localhost/tcp/5001/http ww export ./data/
```

### Export for sharing
```bash
# Export a project directory
ww export ./my-app/

# The resulting IPFS path can be shared with others
# They can then use: ww run /ipfs/QmHash.../my-app
```

## How It Works

1. **Path Resolution**: The command resolves relative paths to absolute paths
2. **File Type Detection**: Determines if the path is a file or directory
3. **Recursive Processing**: For directories, recursively processes all subdirectories and files
4. **IPFS Addition**: Uses the IPFS Unixfs API to add content to the network
5. **Output**: Prints the resulting IPFS path to stdout followed by a newline

## Integration with Other Commands

The exported IPFS paths can be used with other Wetware commands:

```bash
# Export a binary
ww export ./my-program

# Run the exported binary
ww run /ipfs/QmHash.../my-program --help

# Export a project directory
ww export ./my-project/

# Run from the exported directory
ww run /ipfs/QmHash.../my-project/
```

## Error Handling

The command provides clear error messages for common issues:

- **Missing argument**: "export requires exactly one argument: <path>"
- **Path not found**: "path does not exist: /path/to/file"
- **IPFS connection issues**: "failed to add to IPFS: [specific error]"

## Requirements

- IPFS daemon running and accessible
- Proper IPFS API endpoint configuration
- Read access to the specified path

## Environment Variables

- **`WW_IPFS`**: IPFS API endpoint (overrides `--ipfs` flag)

## Security Considerations

- The command only reads files, never modifies them
- Exported content is publicly available on IPFS (unless using private networks)
- Consider what content you're making publicly accessible
- Use appropriate IPFS network configuration for sensitive content

## Troubleshooting

### Connection Issues
```bash
# Check if IPFS daemon is running
ipfs id

# Test IPFS API endpoint
curl http://localhost:5001/api/v0/version
```

### Permission Issues
```bash
# Ensure read access to the path
ls -la /path/to/export

# Check IPFS daemon permissions
ipfs config show | grep "API"
```

### Large File Handling
- Large files are automatically chunked by IPFS
- Progress is not displayed (use `ipfs add` directly for progress)
- Memory usage scales with file size
