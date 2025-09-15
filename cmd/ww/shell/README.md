# Wetware Shell

Production-grade REPL shell built with the Wetware framework and Slurp LISP toolkit. The shell operates in two modes: host mode (spawns a guest process) and guest mode (runs as a cell process).

## Features

- **Dual Mode Operation**: Host mode spawns guest processes, guest mode runs as cell processes
- **REPL**: Built using `github.com/spy16/slurp` for LISP interpretation
- **Readline**: Uses `github.com/chzyer/readline` for terminal experience
- **Wetware Integration**: Runs within Wetware cell environment with system capabilities
- **CLI Integration**: Full urfave/cli integration with configurable options
- **Function Registry**: Extensible set of built-in functions

## Available Functions

### Basic Values
- `nil` - null value
- `true` / `false` - boolean values
- `version` - wetware version string

### Arithmetic
- `(+ a b ...)` - sum of numbers
- `(* a b ...)` - product of numbers
- `(/ a b)` - division (returns float)

### Comparison
- `(= a b)` - equality comparison
- `(> a b)` - greater than
- `(< a b)` - less than

### Utilities
- `help` - display help message
- `println expr` - print expression with newline
- `print expr` - print expression without newline
- `(send "peer-addr-or-id" "proc-id" data)` - send data to a peer process

### IPFS Functions (requires `--with-ipfs` or `--with-all`)
- `(ipfs :cat /ipfs/QmHash/...)` - read IPFS file content as bytes
- `(ipfs :get /ipfs/QmHash/...)` - get IPFS node as file/directory object
- IPFS Path Syntax: `/ipfs/QmHash/...` and `/ipns/domain/...` are automatically parsed

### Execution Functions (requires `--with-exec` or `--with-all`)
- `(exec /ipfs/QmHash/bytecode :timeout 15s)` - execute bytecode from IPFS path

## Usage

The shell is integrated into the main `ww` command:

```bash
# Interactive shell (host mode - spawns guest process)
ww shell

# Execute single command
ww shell -c "(+ 1 2 3)"

# Customize prompt and history
ww shell --prompt "my-shell> " --history-file ~/.ww_history

# Disable banner
ww shell --no-banner

# Environment variable configuration
WW_SHELL_PROMPT="ww> " WW_SHELL_HISTORY="/tmp/ww.tmp" ww shell
```

### Command Line Options

- `-c, --command`: Execute a single command and exit
- `--history-file`: Path to readline history file (default: `/tmp/ww-shell.tmp`)
- `--prompt`: Shell prompt string (default: `"ww> "`)
- `--no-banner`: Disable welcome banner
- `--ipfs`: IPFS API endpoint (default: `/dns4/localhost/tcp/5001/http`)

### Capability Flags

- `--with-ipfs`: Grant IPFS capability (enables `ipfs` functions and path syntax)
- `--with-exec`: Grant process execution capability (enables `exec` function)
- `--with-console`: Grant console output capability
- `--with-all`: Grant all capabilities (equivalent to all above flags)

### Environment Variables

- `WW_SHELL_HISTORY`: Override history file path
- `WW_SHELL_PROMPT`: Override prompt string
- `WW_SHELL_NO_BANNER`: Disable banner (set to any value)
- `WW_IPFS`: Override IPFS API endpoint
- `WW_WITH_IPFS`: Enable IPFS capability
- `WW_WITH_EXEC`: Enable execution capability
- `WW_WITH_CONSOLE`: Enable console capability
- `WW_WITH_ALL`: Enable all capabilities

## Example Session

### Basic Usage
```
Welcome to Wetware Shell! Type 'help' for available commands.
ww> help
Wetware Shell - Available commands:
  help                    - Show this help message
  version                 - Show wetware version
  (+ a b ...)            - Sum numbers
  (* a b ...)            - Multiply numbers
  (= a b)                - Compare equality
  (> a b)                - Greater than
  (< a b)                - Less than
  (println expr)         - Print expression with newline
  (print expr)           - Print expression without newline
  (send "peer-addr-or-id" "proc-id" data) - Send data to a peer process

ww> (+ 1 2 3 4)
10
ww> (* 2 3 4)
24
ww> (> 10 5)
true
ww> (println "Hello, Wetware!")
Hello, Wetware!
ww>
```

### With IPFS Capability
```bash
# Start shell with IPFS capability
ww shell --with-ipfs
```

```
ww> /ipfs/QmHash/example.txt
Path: /ipfs/QmHash/example.txt
ww> (ipfs :cat /ipfs/QmHash/example.txt)
# Returns file content as bytes
ww> (ipfs :get /ipfs/QmHash/example.txt)
<IPFS File: 1024 bytes>
ww> (ipfs :get /ipfs/QmHash/example.txt :read-string)
"Hello from IPFS!"
ww>
```

### With Execution Capability
```bash
# Start shell with execution capability
ww shell --with-exec
```

```
ww> (exec /ipfs/QmHash/bytecode.wasm :timeout 30s)
/protocol/identifier
ww>
```

### With All Capabilities
```bash
# Start shell with all capabilities
ww shell --with-all
```

```
ww> help
# Shows all available functions including ipfs and exec
ww> (send "12D3KooW..." "my-process" "Hello, peer!")
# Sends data to peer process
ww>
```

## Architecture

The shell operates in two distinct modes with capability-based function loading:

### Host Mode
- **Detection**: Runs when `WW_CELL` environment variable is not set
- **Process Spawning**: Uses `ww run -env WW_CELL=true ww -- shell` to spawn guest process
- **Flag Forwarding**: Passes all shell-specific flags and capability flags to the guest process
- **Stdio Forwarding**: Forwards stdin, stdout, and stderr to guest process
- **IPFS Environment**: Initializes IPFS environment before spawning guest process

### Guest Mode (Cell Process)
- **Detection**: Runs when `WW_CELL=true` environment variable is set
- **RPC Connection**: Uses file descriptor 3 for RPC communication with host
- **Wetware Integration**: Connects to Wetware cell system for capabilities
- **Slurp Interpreter**: Provides LISP evaluation engine with IPFS path support
- **Custom REPL**: Production-grade read-eval-print loop with error handling
- **Readline Integration**: Enhanced terminal input with history and completion
- **Function Registry**: Capability-based function loading (base + session-specific)
- **IPFS Integration**: Direct IPFS API access when `--with-ipfs` is enabled
- **Execution Support**: Process execution capability when `--with-exec` is enabled

### Capability System
- **Base Globals**: Always available (arithmetic, comparison, utilities, send)
- **IPFS Capability**: Loads `ipfs` object and enables IPFS path syntax when `--with-ipfs` is set
- **Execution Capability**: Loads `exec` function when `--with-exec` is set
- **Console Capability**: Enables console output when `--with-console` is set
- **All Capabilities**: `--with-all` enables all capabilities at once

### Process Flow
1. `ww shell` (host) → `ww run` (process isolation) → `ww shell` (guest/cell)
2. Host mode initializes IPFS environment and delegates to `ww run` for proper file descriptor management
3. Guest mode runs as a cell process with capability-based function loading
4. Functions are loaded based on capability flags passed from host mode

## Extending

### Adding Base Functions
To add functions that are always available, modify the `globals` map in `globals.go`:

```go
"my-function": slurp.Func("my-function", func(args ...core.Any) {
    // Implementation
    return result
}),
```

### Adding Capability-Based Functions
To add functions that require specific capabilities, modify the `getBaseGlobals()` function in `shell.go`:

```go
// Add IPFS capability function
if c.Bool("with-ipfs") || c.Bool("with-all") {
    gs["my-ipfs-function"] = &MyIPFSFunction{CoreAPI: env.IPFS}
}

// Add execution capability function  
if c.Bool("with-exec") || c.Bool("with-all") {
    gs["my-exec-function"] = &MyExecFunction{Session: session}
}
```

### Adding CLI Flags
To add new CLI flags, modify the `Command()` function in `shell.go`:

```go
&cli.StringFlag{
    Name:    "my-flag",
    Usage:   "description of my flag",
    EnvVars: []string{"WW_MY_FLAG"},
},
```

### Adding New Capabilities
To add a new capability system:

1. Add the capability flag to `flags.go`
2. Add capability check in `getBaseGlobals()` or `NewSessionGlobals()`
3. Implement capability-specific functions
4. Update help message and documentation

## Dependencies

- `github.com/spy16/slurp` - LISP toolkit
- `github.com/chzyer/readline` - Terminal readline support
- `github.com/urfave/cli/v2` - CLI framework
- `capnproto.org/go/capnp/v3` - Cap'n Proto RPC
- `github.com/wetware/go` - Wetware framework
