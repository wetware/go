# Shell Example

Production-grade REPL shell built with the Wetware framework and Slurp LISP toolkit.

## Features

- **REPL**: Built using `github.com/spy16/slurp` for LISP interpretation
- **Readline**: Uses `github.com/chzyer/readline` for terminal experience
- **Wetware Integration**: Runs within Wetware cell environment with system capabilities
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

```bash
# Build
go build -o wetware-shell

# Run (requires wetware environment)
./wetware-shell

# Or run directly
go run main.go
```

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

1. **Wetware Environment**: Integrates with Wetware cell system for capabilities
2. **Slurp Interpreter**: Provides LISP evaluation engine
3. **Custom REPL**: Production-grade read-eval-print loop with error handling
4. **Readline Integration**: Enhanced terminal input with history and completion
5. **Function Registry**: Extensible set of built-in functions

## Extending

To add new functions, modify the `createWetwareEnvironment` function in `main.go`:

```go
"my-function": slurp.Func("my-function", func(args ...core.Any) {
    // Implementation
    return result
}),
```

## Dependencies

- `github.com/spy16/slurp` - LISP toolkit
- `github.com/chzyer/readline` - Terminal readline support
- `capnproto.org/go/capnp/v3` - Cap'n Proto RPC
- `github.com/wetware/go` - Wetware framework
