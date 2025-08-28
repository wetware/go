package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/chzyer/readline"
	"github.com/ipfs/boxo/path"
	"github.com/spy16/slurp"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"

	"github.com/spy16/slurp/repl"
	"github.com/wetware/go/system"
)

var config = readline.Config{
	Prompt:          "ww> ",
	HistoryFile:     "/tmp/ww-shell.tmp",
	AutoComplete:    getCompleter(),
	InterruptPrompt: "^C",
	EOFPrompt:       "exit",
}

func main() {
	ctx := context.Background()

	// Check command line arguments for -c flag
	if len(os.Args) > 1 && os.Args[1] == "-c" {
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: %s -c <command>\n", os.Args[0])
			os.Exit(1)
		}

		// Execute single command
		if err := executeCommand(ctx, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "repl error: %s\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if the bootstrap file descriptor exists
	host := os.NewFile(system.BOOTSTRAP_FD, "host")
	if host == nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create bootstrap file descriptor\n")
		os.Exit(1)
	}

	conn := rpc.NewConn(rpc.NewStreamTransport(host), &rpc.Options{
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

func runREPL(ctx context.Context, i system.Importer) error {
	// Create readline instance
	rl, err := readline.NewEx(&config)
	if err != nil {
		return err
	}
	defer rl.Close()

	// Create the REPL with custom options using our custom reader factory
	return repl.New(newInterpreter(i),
		repl.WithBanner("Welcome to Wetware Shell! Type 'help' for available commands."),
		repl.WithPrompts("ww ", "  | "),
		repl.WithPrinter(printer{out: os.Stdout}),
		repl.WithReaderFactory(DefaultReaderFactory{}),
		repl.WithInput(lineReader{Driver: rl}, func(err error) error {
			if err == nil || err == readline.ErrInterrupt {
				return nil
			}
			return err
		}),
	).Loop(ctx)
}

// executeCommand executes a single command line
func executeCommand(ctx context.Context, command string) error {
	// Create a basic interpreter without import functionality for testing
	env := slurp.New()

	// Add basic globals for testing
	globals := map[string]core.Any{
		"nil":     builtin.Nil{},
		"true":    builtin.Bool(true),
		"false":   builtin.Bool(false),
		"version": builtin.String("wetware-0.1.0"),
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

	if err := env.Bind(globals); err != nil {
		return fmt.Errorf("failed to bind globals: %w", err)
	}

	// Create a reader from the command string
	commandReader := strings.NewReader(command)

	// Create the REPL with the command reader
	return repl.New(env,
		repl.WithBanner(""),      // No banner for single command execution
		repl.WithPrompts("", ""), // No prompts for single command execution
		repl.WithPrinter(printer{out: os.Stdout}),
		repl.WithReaderFactory(DefaultReaderFactory{}),
		repl.WithInput(commandReaderWrapper{Reader: commandReader}, func(err error) error {
			if err != nil && err != io.EOF {
				return err
			}
			return nil
		}),
	).Loop(ctx)
}

// commandReaderWrapper implements the repl.Input interface for single command execution
type commandReaderWrapper struct {
	*strings.Reader
}

func (r commandReaderWrapper) Readline() (string, error) {
	// Read a line from the strings.Reader
	buf := make([]byte, 0, 1024)
	for {
		if b, err := r.ReadByte(); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		} else if b == '\n' {
			break
		} else {
			buf = append(buf, b)
		}
	}
	return strings.TrimSpace(string(buf)), nil
}

func (r commandReaderWrapper) Prompt(p string) {
	// No prompting for single command execution
}

// Path represents an IPFS path and implements core.Any
type Path struct {
	path.Path
}

func (p Path) String() string {
	return p.Path.String()
}

// newInterpreter creates a slurp environment with wetware-specific functions
func newInterpreter(i system.Importer) *slurp.Interpreter {
	// Create analyzer for special forms
	analyzer := &builtin.Analyzer{
		Specials: make(map[string]builtin.ParseSpecial),
	}

	// Only add import special form if importer is available
	if i.IsValid() {
		analyzer.Specials["import"] = func(analyzer core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
			expr := builtin.DoExpr{}
			err := core.ForEach(args, func(item core.Any) (bool, error) {
				if e, ok := item.(system.ServiceToken); ok {
					expr = append(expr, ImportExpr{Client: i, ServiceToken: e})
					return true, nil
				}
				return false, fmt.Errorf("expected envelope, got %T", item)
			})
			return expr, err
		}
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
  (import "module")      - Import a module (stubbed)

  IPFS Path Syntax:
  /ipfs/QmHash/...       - Direct IPFS path
  /ipns/domain/...       - IPNS path`
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

// printer implements the repl.Printer interface for better output formatting
type printer struct {
	out io.Writer
}

func (p printer) Print(val interface{}) error {
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
	case Path:
		_, err := fmt.Fprintf(p.out, "Path: %s\n", v.String())
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
