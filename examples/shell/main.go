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

	// Create a custom environment with wetware-specific functions
	env := newInterpreter(client)
	// if err := env.Bind(globals); err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to bind globals: %v\n", err)
	// 	os.Exit(1)
	// }

	// Create a production-grade REPL with readline support
	r := createProductionREPL(env)

	// Run the REPL until context is cancelled
	if err := r.Loop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "repl error: %s\n", err)
		os.Exit(1)
	}
}

// newInterpreter creates a slurp environment with wetware-specific functions
func newInterpreter(client capnp.Client) *slurp.Interpreter {
	// Create base environment
	env := slurp.New()

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
  (print expr)           - Print expression without newline`
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

// createProductionREPL creates a production-grade REPL with readline support
func createProductionREPL(env *slurp.Interpreter) *repl.REPL {
	// Create readline instance
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "wetware> ",
		HistoryFile:     "/tmp/wetware_shell.tmp",
		AutoComplete:    getCompleter(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create readline: %v\n", err)
		// Fallback to basic REPL
		return repl.New(env,
			repl.WithBanner("Welcome to Wetware Shell!"),
			repl.WithPrompts("wetware> ", "  | "))
	}

	// Create the REPL with custom options
	r := repl.New(env,
		repl.WithBanner("Welcome to Wetware Shell! Type 'help' for available commands."),
		repl.WithPrompts("ww ", "  | "),
		repl.WithPrinter(&customPrinter{out: os.Stdout}),
		repl.WithInput(stdio{Driver: rl}, func(err error) error {
			if err == nil || err == readline.ErrInterrupt {
				return nil
			}
			// Close readline when we're done
			if err == io.EOF {
				rl.Close()
			}
			return err
		}),
	)

	return r
}

// customPrinter implements the repl.Printer interface for better output formatting
type customPrinter struct {
	out io.Writer
}

func (p *customPrinter) Print(val interface{}) error {
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

// stdio implements the repl.Input interface using readline
type stdio struct {
	Driver *readline.Instance
}

func (s stdio) Readline() (string, error) {
	line, err := s.Driver.Readline()
	return line, err
}

// Prompt implements the repl.Prompter interface
func (s stdio) Prompt(p string) {
	s.Driver.SetPrompt(p)
}

// getCompleter returns a readline completer for wetware commands
func getCompleter() readline.AutoCompleter {
	return readline.NewPrefixCompleter(
		readline.PcItem("help"),
		readline.PcItem("version"),
		readline.PcItem("println"),
		readline.PcItem("print"),
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
