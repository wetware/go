package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/chzyer/readline"
	"github.com/spy16/slurp"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/repl"
	"github.com/wetware/go/system"
)

var config = readline.Config{
	Prompt:          "ww ",
	HistoryFile:     "/tmp/ww-shell.tmp",
	AutoComplete:    getCompleter(),
	InterruptPrompt: "^C",
	EOFPrompt:       "exit",
}

func main() {
	ctx := context.Background()

	// Check if the bootstrap file descriptor exists
	bootstrapFile := os.NewFile(system.BOOTSTRAP_FD, "host")
	if bootstrapFile == nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create bootstrap file descriptor\n")
		os.Exit(1)
	}

	conn := rpc.NewConn(rpc.NewStreamTransport(bootstrapFile), &rpc.Options{
		BaseContext: func() context.Context { return ctx },
		// BootstrapClient: export(),
	})
	defer conn.Close()

	client := conn.Bootstrap(ctx)
	defer client.Release()

	if err := runREPL(ctx, system.Importer(client)); err != nil {
		fmt.Fprintf(os.Stderr, "repl error: %s\n", err)
		os.Exit(1)
	}
}

func runREPL[T ~capnp.ClientKind](ctx context.Context, t T) error {
	// Create readline instance
	rl, err := readline.NewEx(&config)
	if err != nil {
		return err
	}
	defer rl.Close()

	// Create the REPL with custom options
	return repl.New(newInterpreter(t),
		repl.WithBanner("Welcome to Wetware Shell! Type 'help' for available commands."),
		repl.WithPrompts("ww ", "  | "),
		repl.WithPrinter(&printer{out: os.Stdout}),
		repl.WithInput(lineReader{Driver: rl}, func(err error) error {
			if err == nil || err == readline.ErrInterrupt {
				return nil
			}
			return err
		}),
	).Loop(ctx)
}

// newInterpreter creates a slurp environment with wetware-specific functions
func newInterpreter[T ~capnp.ClientKind](t T) *slurp.Interpreter {
	// Create analyzer for special forms
	analyzer := &builtin.Analyzer{
		Specials: map[string]builtin.ParseSpecial{
			"import": func(analyzer core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
				// Special form: (import "module-name")
				// For now, just return a placeholder message
				return &ImportExpr[T]{Client: t, MethodArgs: args}, nil
			},
		},
	}

	// Create base environment with analyzer
	env := slurp.New(slurp.WithAnalyzer(analyzer))

	// Add wetware-specific globals
	globals := map[string]core.Any{
		// Basic values
		"nil":     builtin.Nil{},
		"true":    builtin.Bool(true),
		"false":   builtin.Bool(false),
		"version": builtin.String("wetware-0.1.0"),

		// Basic operations
		"=": slurp.Func("=", core.Eq),
		"+": slurp.Func("sum", func(a ...int) int {
			sum := 0
			for _, item := range a {
				sum += item
			}
			return sum
		}),
		">": slurp.Func(">", func(a, b builtin.Int64) bool {
			return a > b
		}),
		"<": slurp.Func("<", func(a, b builtin.Int64) bool {
			return a < b
		}),
		"*": slurp.Func("*", func(a ...int) int {
			product := 1
			for _, item := range a {
				product *= item
			}
			return product
		}),
		"/": slurp.Func("/", func(a, b builtin.Int64) float64 {
			return float64(a) / float64(b)
		}),

		// Wetware-specific functions
		"help": slurp.Func("help", func() string {
			return `Wetware Shell - Available commands:
  help                    - Show this help message
  version                 - Show wetware version
  (+ a b ...)            - Sum numbers
  (* a b ...)            - Multiply numbers
  (= a b)                - Compare equality
  (> a b)                - Greater than
  (< a b)                - Less than
  (println expr)         - Print expression with newline
  (print expr)           - Print expression without newline
  (import "module")      - Import a module (stubbed)`
		}),
		"println": slurp.Func("println", func(args ...core.Any) {
			for _, arg := range args {
				fmt.Println(arg)
			}
		}),
		"print": slurp.Func("print", func(args ...core.Any) {
			for _, arg := range args {
				fmt.Print(arg)
			}
		}),
	}

	// Bind the globals to the environment
	if err := env.Bind(globals); err != nil {
		fmt.Fprintf(os.Stderr, "failed to bind globals: %v\n", err)
	}

	return env
}

// ImportExpr implements core.Expr for import statements
type ImportExpr[T ~capnp.ClientKind] struct {
	Client     T
	MethodArgs core.Seq
}

func (e ImportExpr[T]) Eval(env core.Env) (core.Any, error) {
	// Parse the import arguments to get the module name
	// For now, we'll assume a simple string argument
	// TODO: Handle more complex import syntax if needed

	// Convert the method args to a slice to access the first argument
	argsSlice, err := core.ToSlice(e.MethodArgs)
	if err != nil {
		return nil, fmt.Errorf("import: failed to parse arguments: %w", err)
	}

	if len(argsSlice) == 0 {
		return nil, fmt.Errorf("import: missing module name")
	}

	// Get the module name (first argument)
	moduleName := argsSlice[0]

	// Convert module name to string
	var moduleNameStr string
	switch v := moduleName.(type) {
	case builtin.String:
		moduleNameStr = string(v)
	case builtin.Symbol:
		moduleNameStr = string(v)
	default:
		return nil, fmt.Errorf("import: module name must be a string or symbol, got %T", moduleName)
	}

	// For now, return a message showing what we would import
	// TODO: Actually call the system.Importer.Import method
	return builtin.String(fmt.Sprintf("import: would import module '%s' using system capability", moduleNameStr)), nil
}

// printer implements the repl.Printer interface for better output formatting
type printer struct {
	out io.Writer
}

func (p *printer) Print(val interface{}) error {
	switch v := val.(type) {
	case nil:
		_, err := fmt.Fprintf(p.out, "nil\n")
		return err
	case builtin.Bool:
		_, err := fmt.Fprintf(p.out, "%t\n", bool(v))
		return err
	case builtin.Int64:
		_, err := fmt.Fprintf(p.out, "%d\n", int64(v))
		return err
	case builtin.String:
		_, err := fmt.Fprintf(p.out, "%s\n", string(v))
		return err
	case builtin.Float64:
		_, err := fmt.Fprintf(p.out, "%g\n", float64(v))
		return err
	case builtin.Nil:
		_, err := fmt.Fprintf(p.out, "nil\n")
		return err
	default:
		// For any other type, use Go's default formatting
		_, err := fmt.Fprintf(p.out, "%v\n", v)
		return err
	}
}

// lineReader implements the repl.Input interface using readline
type lineReader struct {
	Driver *readline.Instance
}

func (s lineReader) Readline() (string, error) {
	line, err := s.Driver.Readline()
	return line, err
}

// Prompt implements the repl.Prompter interface
func (s lineReader) Prompt(p string) {
	s.Driver.SetPrompt(p)
}

// getCompleter returns a readline completer for wetware commands
func getCompleter() readline.AutoCompleter {
	return readline.NewPrefixCompleter(
		readline.PcItem("help"),
		readline.PcItem("version"),
		readline.PcItem("println"),
		readline.PcItem("print"),
		readline.PcItem("system"),
		readline.PcItem("callc"),
		readline.PcItem("+"),
		readline.PcItem("*"),
		readline.PcItem("="),
		readline.PcItem(">"),
		readline.PcItem("<"),
		readline.PcItem("nil"),
		readline.PcItem("true"),
		readline.PcItem("false"),
	)
}
