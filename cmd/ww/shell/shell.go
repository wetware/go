//go:generate capnp compile -I.. -I$GOPATH/src/capnproto.org/go/capnp/std -ogo shell.capnp

package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/chzyer/readline"
	"github.com/spy16/slurp"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/urfave/cli/v2"

	"github.com/spy16/slurp/repl"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
	"github.com/wetware/go/util"
)

var env util.IPFSEnv

func getReadlineConfig(c *cli.Context) readline.Config {
	return readline.Config{
		Prompt:          c.String("prompt"),
		HistoryFile:     c.String("history-file"),
		AutoComplete:    getCompleter(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	}
}

func Command() *cli.Command {
	return &cli.Command{
		Name: "shell",
		Before: func(c *cli.Context) error {
			addr := c.String("ipfs")
			return env.Boot(addr)
		},
		After: func(c *cli.Context) error {
			return env.Close()
		},
		Action: Main,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Value:   "/dns4/localhost/tcp/5001/http",
			},
			&cli.StringFlag{
				Name:    "command",
				Aliases: []string{"c"},
				Usage:   "execute a single command and exit",
			},
			&cli.StringFlag{
				Name:    "history-file",
				Usage:   "path to readline history file",
				Value:   "/tmp/ww-shell.tmp",
				EnvVars: []string{"WW_SHELL_HISTORY"},
			},
			&cli.StringFlag{
				Name:    "prompt",
				Usage:   "shell prompt string",
				Value:   "ww> ",
				EnvVars: []string{"WW_SHELL_PROMPT"},
			},
			&cli.BoolFlag{
				Name:    "no-banner",
				Usage:   "disable welcome banner",
				EnvVars: []string{"WW_SHELL_NO_BANNER"},
			},
		},
	}
}

func Main(c *cli.Context) error {
	ctx := c.Context

	// Check if we're in guest mode (cell process)
	if os.Getenv("WW_CELL") == "true" {
		return runGuestMode(ctx, c)
	}

	// Host mode: spawn guest process with ww run
	return runHostMode(ctx, c)
}

// runHostMode runs the shell in host mode, spawning a guest process
func runHostMode(ctx context.Context, c *cli.Context) error {
	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build the command to run the shell in guest mode
	cmd := exec.CommandContext(ctx, execPath, "run", "-env", "WW_CELL=true", execPath, "--", "shell")

	// Pass through shell-specific flags
	if command := c.String("command"); command != "" {
		cmd.Args = append(cmd.Args, "-c", command)
	}
	if historyFile := c.String("history-file"); historyFile != "" {
		cmd.Args = append(cmd.Args, "--history-file", historyFile)
	}
	if prompt := c.String("prompt"); prompt != "" {
		cmd.Args = append(cmd.Args, "--prompt", prompt)
	}
	if c.Bool("no-banner") {
		cmd.Args = append(cmd.Args, "--no-banner")
	}

	// Set up stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	return cmd.Run()
}

// runGuestMode runs the shell in guest mode (as a cell process)
func runGuestMode(ctx context.Context, c *cli.Context) error {
	// Check if command flag is provided
	if command := c.String("command"); command != "" {
		// Execute single command
		if err := executeCommand(ctx, command); err != nil {
			return fmt.Errorf("repl error: %w", err)
		}
		return nil
	}

	// Check if the bootstrap file descriptor exists
	host := os.NewFile(system.BOOTSTRAP_FD, "host")
	if host == nil {
		return fmt.Errorf("failed to create bootstrap file descriptor")
	}

	conn := rpc.NewConn(rpc.NewStreamTransport(host), &rpc.Options{
		BaseContext: func() context.Context { return ctx },
		// BootstrapClient: export(),
	})
	defer conn.Close()

	client := conn.Bootstrap(ctx)
	defer client.Release()

	f, release := system.Terminal(client).Login(ctx, nil)
	defer release()

	// Create base environment with analyzer and globals to it.
	eval := slurp.New()

	// Bind base globals (common to both modes)
	if err := eval.Bind(getBaseGlobals(ctx)); err != nil {
		return fmt.Errorf("failed to bind base globals: %w", err)
	}

	// Bind session-specific globals (interactive mode only)
	if err := eval.Bind(getSessionGlobals(ctx, f)); err != nil {
		return fmt.Errorf("failed to bind session globals: %w", err)
	}

	// Set up readline
	config := getReadlineConfig(c)
	rl, err := readline.NewEx(&config)
	if err != nil {
		return fmt.Errorf("readline: %w", err)
	}
	defer rl.Close()

	// Configure banner
	var banner string
	if !c.Bool("no-banner") {
		banner = "Welcome to Wetware Shell! Type 'help' for available commands."
	}

	if err := repl.New(eval,
		repl.WithBanner(banner),
		repl.WithPrompts("ww ", "  | "),
		repl.WithPrinter(printer{out: os.Stdout}),
		repl.WithReaderFactory(lang.DefaultReaderFactory{IPFS: env.IPFS}),
		repl.WithInput(lineReader{Driver: rl}, func(err error) error {
			if err == nil || err == readline.ErrInterrupt {
				return nil
			}
			return err
		}),
	).Loop(ctx); err != nil {
		return fmt.Errorf("repl: %w", err)
	}
	return nil
}

// executeCommand executes a single command line
func executeCommand(ctx context.Context, command string) error {
	// Use the global IPFS environment that was already initialized

	// Create a basic interpreter without import functionality for testing
	eval := slurp.New()

	// Use the same base globals as the interactive mode
	if err := eval.Bind(getBaseGlobals(ctx)); err != nil {
		return fmt.Errorf("failed to bind globals: %w", err)
	}

	// Create a reader from the command string
	commandReader := strings.NewReader(command)

	// Create a reader factory for IPFS path support
	readerFactory := lang.DefaultReaderFactory{IPFS: env.IPFS}

	// Read and evaluate the command directly
	reader := readerFactory.NewReader(commandReader)

	// Read the expression
	expr, err := reader.One()
	if err != nil {
		if err == io.EOF {
			return nil // Empty command, nothing to do
		}
		return fmt.Errorf("failed to read command: %w", err)
	}

	// Evaluate the expression
	result, err := eval.Eval(expr)
	if err != nil {
		return fmt.Errorf("failed to evaluate command: %w", err)
	}

	// Print the result if it's not nil
	if result != nil {
		printer := printer{out: os.Stdout}
		return printer.Print(result)
	}

	return nil
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
	case *lang.Path:
		_, err := fmt.Fprintf(p.out, "Path: %s\n", v.Path.String())
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
		readline.PcItem("exec"),
		readline.PcItem("ipfs"),
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

// getBaseGlobals returns the base globals that are common to both interactive and command modes
func getBaseGlobals(ctx context.Context) map[string]core.Any {
	baseGlobals := make(map[string]core.Any)

	// Copy the base globals from globals.go
	for k, v := range globals {
		baseGlobals[k] = v
	}

	// Add IPFS support (available in both modes)
	baseGlobals["ipfs"] = lang.NewIPFS(ctx, env.IPFS)

	return baseGlobals
}

// getSessionGlobals returns additional globals for interactive mode (requires terminal connection)
func getSessionGlobals(ctx context.Context, f system.Terminal_login_Results_Future) map[string]core.Any {
	sessionGlobals := make(map[string]core.Any)

	// Add exec functionality (only available in interactive mode with terminal connection)
	sessionGlobals["exec"] = lang.NewExecutor(ctx, f.Exec())

	return sessionGlobals
}
