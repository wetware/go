# ww cat

The `ww cat` command connects your local stdin/stdout to a remote peer's stream, allowing you to interact with remote WASM processes as if they were local.

## Usage

```bash
ww cat <peer-id> <endpoint>
```

### Arguments

- `<peer-id>`: The libp2p peer ID of the remote peer
- `<endpoint>`: The base58-encoded endpoint identifier

### Options

- `--ipfs <address>`: IPFS API endpoint (default: "/dns4/localhost/tcp/5001/http")
- `--port <port>`: Local port for P2P networking (default: 2020)
- `--with-p2p`: Enable P2P networking capability (required for cat)

## Examples

### Basic usage

```bash
# Connect to a remote echo process
echo "hello, remote!" | ww cat 12D3KooW... EYEHuCPU8RX

# Send a file to a remote process
cat myfile.txt | ww cat 12D3KooW... EYEHuCPU8RX

# Receive output from a remote process
ww cat 12D3KooW... EYEHuCPU8RX > output.txt
```

### With capabilities

```bash
# Enable P2P networking (required)
ww cat --with-p2p 12D3KooW... EYEHuCPU8RX

# Enable all capabilities
ww cat --with-all 12D3KooW... EYEHuCPU8RX
```

## How it works

1. Parses the peer ID and base58-encoded endpoint
2. Constructs the full protocol ID: `/ww/0.1.0/<endpoint>`
3. Opens a libp2p stream to the remote peer
4. Sets up bidirectional copying:
   - Local stdin → Remote stream
   - Remote stream → Local stdout
5. Handles stream closure and errors gracefully

## Use cases

- **Remote debugging**: Connect to a remote WASM process for debugging
- **Data processing**: Send data to remote processes for processing
- **Interactive sessions**: Have interactive sessions with remote processes
- **File transfer**: Transfer files through remote processes
- **Distributed computing**: Use remote processes as compute nodes

## Requirements

- The remote peer must be running a WASM process with the specified endpoint
- P2P networking capability must be enabled (`--with-p2p`)
- The remote peer must be reachable via libp2p
