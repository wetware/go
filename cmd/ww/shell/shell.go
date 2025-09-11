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
	"github.com/ipfs/boxo/path"
	"github.com/spy16/slurp"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/urfave/cli/v2"

	"github.com/spy16/slurp/repl"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
)

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
		Name:   "shell",
		Action: Main,
		Flags: []cli.Flag{
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
	env := slurp.New()
	if err := env.Bind(globals); err != nil {
		return fmt.Errorf("failed to bind globals: %w", err)
	}
	if err := env.Bind(session(ctx, f)); err != nil {
		return fmt.Errorf("failed to bind session: %w", err)
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

	if err := repl.New(env,
		repl.WithBanner(banner),
		repl.WithPrompts("ww ", "  | "),
		repl.WithPrinter(printer{out: os.Stdout}),
		repl.WithReaderFactory(DefaultReaderFactory{}),
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

func session(ctx context.Context, f system.Terminal_login_Results_Future) map[string]core.Any {
	return map[string]core.Any{
		"exec": lang.NewExecutor(ctx, f.Exec()),
	}
}
