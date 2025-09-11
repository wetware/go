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

### Environment Variables

- `WW_SHELL_HISTORY`: Override history file path
- `WW_SHELL_PROMPT`: Override prompt string
- `WW_SHELL_NO_BANNER`: Disable banner (set to any value)

## Example Session

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

## Architecture

The shell operates in two distinct modes:

### Host Mode
- **Detection**: Runs when `WW_CELL` environment variable is not set
- **Process Spawning**: Uses `ww run -env WW_CELL=true ww -- shell` to spawn guest process
- **Flag Forwarding**: Passes all shell-specific flags to the guest process
- **Stdio Forwarding**: Forwards stdin, stdout, and stderr to guest process

### Guest Mode (Cell Process)
- **Detection**: Runs when `WW_CELL=true` environment variable is set
- **RPC Connection**: Uses file descriptor 3 for RPC communication with host
- **Wetware Integration**: Connects to Wetware cell system for capabilities
- **Slurp Interpreter**: Provides LISP evaluation engine
- **Custom REPL**: Production-grade read-eval-print loop with error handling
- **Readline Integration**: Enhanced terminal input with history and completion
- **Function Registry**: Extensible set of built-in functions

### Process Flow
1. `ww shell` (host) → `ww run` (process isolation) → `ww shell` (guest/cell)
2. Host mode delegates to `ww run` for proper file descriptor management
3. Guest mode runs as a cell process with full Wetware capabilities

## Extending

To add new functions, modify the `globals` map in `globals.go`:

```go
"my-function": slurp.Func("my-function", func(args ...core.Any) {
    // Implementation
    return result
}),
```

To add new CLI flags, modify the `Command()` function in `shell.go`:

```go
&cli.StringFlag{
    Name:    "my-flag",
    Usage:   "description of my flag",
    EnvVars: []string{"WW_MY_FLAG"},
},
```

## Dependencies

- `github.com/spy16/slurp` - LISP toolkit
- `github.com/chzyer/readline` - Terminal readline support
- `github.com/urfave/cli/v2` - CLI framework
- `capnproto.org/go/capnp/v3` - Cap'n Proto RPC
- `github.com/wetware/go` - Wetware framework
