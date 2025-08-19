# Wetware Import Command

The `ww import` command downloads files and directories from IPFS to the local filesystem, providing the inverse functionality of the `ww export` command.

## Overview

The import command enables users to retrieve content from IPFS and save it locally. This is particularly useful for:
- Downloading binaries published via `ww export`
- Retrieving project files and directories from IPFS
- Creating local copies of IPFS content for development or execution
- Complementing the export command for a complete IPFS workflow

## Usage

```bash
ww import <ipfs-path> [local-path]
```

## Arguments

- **`<ipfs-path>`**: The IPFS path to import from (required)
  - Must be a valid IPFS path (e.g., `/ipfs/QmHash...`)
  - Supports both files and directories
  - Can include subpaths (e.g., `/ipfs/QmHash.../subdirectory/file.txt`)

- **`[local-path]`**: The local destination path (optional)
  - Defaults to current directory (`.`) if not specified
  - Can be a file path or directory path
  - If directory, content is placed inside with original names
  - If file path, content is saved with that exact name

## Options

- **`--ipfs`**: IPFS API endpoint (default: `/dns4/localhost/tcp/5001/http`)
  - Can also be set via `WW_IPFS` environment variable
  - Supports multiaddr, URL, or local path formats

- **`--executable, -x`**: Make imported files executable
  - Applies `chmod +x` to all imported files
  - Useful for importing binaries and scripts
  - Default: false

## Examples

### Import a single file
```bash
# Import to current directory with original filename
ww import /ipfs/QmHash.../document.txt

# Import to specific file path
ww import /ipfs/QmHash.../document.txt ./local_doc.txt

# Import binary and make executable
ww import /ipfs/QmHash.../my-program --executable
```

### Import a directory
```bash
# Import entire directory to current location
ww import /ipfs/QmHash.../my-project

# Import to specific directory
ww import /ipfs/QmHash.../my-project ./downloaded-project

# Import with executable permissions
ww import /ipfs/QmHash.../my-project --executable
```

### Import with custom IPFS endpoint
```bash
# Use custom IPFS endpoint
ww import --ipfs=/ip4/127.0.0.1/tcp/5001/http /ipfs/QmHash.../file.txt

# Use environment variable
WW_IPFS=/dns4/localhost/tcp/5001/http ww import /ipfs/QmHash.../file.txt
```

## How It Works

1. **IPFS Path Validation**: Parses and validates the provided IPFS path
2. **Content Retrieval**: Downloads the content from IPFS using the Unixfs API
3. **Type Detection**: Determines if the content is a file or directory
4. **Local Path Resolution**: Resolves the local destination path
5. **Content Extraction**: 
   - For files: Downloads and saves to the specified location
   - For directories: Recursively extracts all files and subdirectories
6. **Permission Setting**: Applies executable permissions if `--executable` flag is used

## File vs Directory Handling

### Single Files
- Content is downloaded directly to the target location
- If local path is a directory, uses the original filename
- If local path is a file path, uses that exact name
- Parent directories are created automatically

### Directories
- All files and subdirectories are extracted recursively
- Maintains the original directory structure
- Files are placed in their corresponding subdirectories
- Empty directories are preserved

## Integration with Other Commands

### Export → Import Workflow
```bash
# Export a project to IPFS
ww export ./my-project
# Output: /ipfs/QmHash...

# Import the project elsewhere
ww import /ipfs/QmHash... ./downloaded-project
```

### Run Command Integration
The `ww run` command now uses the import functionality internally:
```bash
# Run a binary directly from IPFS
ww run /ipfs/QmHash.../my-program

# This internally uses import to download and make executable
```

## Error Handling

The command provides clear error messages for common issues:

- **Invalid IPFS path**: "invalid IPFS path: [error details]"
- **IPFS connection issues**: "failed to get IPFS path: [specific error]"
- **Local filesystem issues**: "failed to create directory: [specific error]"
- **Permission issues**: "failed to make file executable: [specific error]"

## Requirements

- IPFS daemon running and accessible
- Proper IPFS API endpoint configuration
- Write access to the local destination path
- Sufficient disk space for the imported content

## Environment Variables

- **`WW_IPFS`**: IPFS API endpoint (overrides `--ipfs` flag)

## Security Considerations

- Content is downloaded from IPFS network (verify source authenticity)
- Local files are created with standard permissions (644 for files, 755 for directories)
- Executable flag can make files runnable (use with caution)
- No automatic verification of downloaded content integrity

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
# Ensure write access to destination
ls -la /path/to/destination

# Check IPFS daemon permissions
ipfs config show | grep "API"
```

### Large File Handling
- Large files are downloaded in chunks by IPFS
- Progress is not displayed (use `ipfs get` directly for progress)
- Memory usage scales with file size
- Consider available disk space for large imports

## Use Cases

### Development Workflow
```bash
# Export development environment
ww export ./dev-env

# Import on another machine
ww import /ipfs/QmHash... ./dev-env
```

### Binary Distribution
```bash
# Export compiled binary
ww export ./my-app

# Import and run elsewhere
ww import /ipfs/QmHash... --executable
./my-app
```

### Project Sharing
```bash
# Export project with dependencies
ww export ./my-project

# Share IPFS path with team
# Team members can import with:
ww import /ipfs/QmHash... ./my-project
```

## Future Enhancements

- Progress indicators for large downloads
- Content verification and integrity checking
- Selective file/directory import
- Compression and optimization options
- Integration with IPNS for mutable references
