# Import Command

Downloads files and directories from IPFS to the local filesystem.

## Usage

```bash
ww import <ipfs-path> [local-path]
```

## Arguments

- `<ipfs-path>` - IPFS path to import from (required)
- `[local-path]` - Local destination path (defaults to current directory)

## Options

- `--ipfs` - IPFS API endpoint (default: `/dns4/localhost/tcp/5001/http`)
- `--executable, -x` - Make imported files executable (`chmod +x`)
- `WW_IPFS` environment variable overrides `--ipfs` flag

## Implementation

1. Parses and validates IPFS path
2. Downloads content using IPFS Unixfs API
3. Determines content type (file or directory)
4. Extracts content to local filesystem
5. Applies executable permissions if requested

## Examples

```bash
# Import to current directory
ww import /ipfs/QmHash.../file.txt

# Import to specific path
ww import /ipfs/QmHash.../directory ./local-dir

# Import and make executable
ww import /ipfs/QmHash.../binary --executable
```

## File vs Directory Handling

- **Files**: Downloaded directly to target location
- **Directories**: Recursively extracts all files and subdirectories
- Maintains original directory structure
- Creates parent directories automatically

## Requirements

- IPFS daemon running and accessible
- Write access to local destination path
- Sufficient disk space for imported content

## Error Handling

- Invalid IPFS path: "invalid IPFS path: <error>"
- IPFS errors: "failed to get IPFS path: <error>"
- Filesystem errors: "failed to create directory: <error>"
- Permission errors: "failed to make file executable: <error>"
