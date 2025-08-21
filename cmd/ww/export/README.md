# Export Command

Adds files and directories to IPFS recursively.

## Usage

```bash
ww export <path>
```

## Arguments

- `<path>` - Local filesystem path to export (file or directory)

## Options

- `--ipfs` - IPFS API endpoint (default: `/dns4/localhost/tcp/5001/http`)
- `WW_IPFS` environment variable overrides `--ipfs` flag

## Implementation

1. Resolves relative paths to absolute paths
2. Validates path existence
3. Uses IPFS Unixfs API to add content recursively
4. Outputs IPFS path to stdout followed by newline

## Examples

```bash
# Export current directory
ww export .

# Export specific file
ww export document.txt

# Export with custom IPFS endpoint
ww export --ipfs=/ip4/127.0.0.1/tcp/5001/http ./data/
```

## Requirements

- IPFS daemon running and accessible
- Read access to specified path
- Proper IPFS API endpoint configuration

## Error Handling

- Missing argument: "export requires exactly one argument: <path>"
- Path not found: "path does not exist: <path>"
- IPFS errors: "failed to add to IPFS: <error>"
