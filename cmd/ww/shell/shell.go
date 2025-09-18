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
	"github.com/wetware/go/cmd/internal/flags"
	"github.com/wetware/go/system"
	"github.com/wetware/go/util"
)

var env util.IPFSEnv

func Command() *cli.Command {
	return &cli.Command{
		Name: "shell",
		Before: func(c *cli.Context) error {
			addr := c.String("ipfs")
			if err := env.Boot(addr); err != nil {
				return fmt.Errorf("failed to boot IPFS environment: %w", err)
			}
			return nil
		},
		After: func(c *cli.Context) error {
			return env.Close()
		},
		Action: Main,
		Flags: append([]cli.Flag{
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
		}, flags.CapabilityFlags()...),
	}
}

func Main(c *cli.Context) error {
	// Check if we're in guest mode (cell process)
	if os.Getenv("WW_CELL") == "true" {
		return runGuestMode(c)
	}

	// Host mode: spawn guest process with ww run
	return runHostMode(c)
}

// runHostMode runs the shell in host mode, spawning a guest process
func runHostMode(c *cli.Context) error {
	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build the command to run the shell in guest mode
	cmd := exec.CommandContext(c.Context, execPath, "run", "-env", "WW_CELL=true")

	// Pass through capability flags to the run command
	if c.Bool("with-ipfs") {
		cmd.Args = append(cmd.Args, "--with-ipfs")
	}
	if c.Bool("with-exec") {
		cmd.Args = append(cmd.Args, "--with-exec")
	}
	if c.Bool("with-console") {
		cmd.Args = append(cmd.Args, "--with-console")
	}
	if c.Bool("with-p2p") {
		cmd.Args = append(cmd.Args, "--with-p2p")
	}
	if c.Bool("with-all") {
		cmd.Args = append(cmd.Args, "--with-all")
	}

	// Pass through mDNS capability to the run command
	if c.Bool("with-mdns") {
		cmd.Args = append(cmd.Args, "--with-mdns")
	}

	// Add the executable and shell command
	cmd.Args = append(cmd.Args, execPath, "--", "shell")

	// Pass through capability flags to the shell command as well
	if c.Bool("with-ipfs") {
		cmd.Args = append(cmd.Args, "--with-ipfs")
	}
	if c.Bool("with-exec") {
		cmd.Args = append(cmd.Args, "--with-exec")
	}
	if c.Bool("with-console") {
		cmd.Args = append(cmd.Args, "--with-console")
	}
	if c.Bool("with-mdns") {
		cmd.Args = append(cmd.Args, "--with-mdns")
	}
	if c.Bool("with-p2p") {
		cmd.Args = append(cmd.Args, "--with-p2p")
	}
	if c.Bool("with-all") {
		cmd.Args = append(cmd.Args, "--with-all")
	}

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
func runGuestMode(c *cli.Context) error {
	// Check if the bootstrap file descriptor exists
	host := os.NewFile(system.BOOTSTRAP_FD, "host")
	if host == nil {
		return fmt.Errorf("failed to create bootstrap file descriptor")
	}

	// Check if command flag is provided
	if command := c.String("command"); command != "" {
		// Execute single command
		if err := executeCommand(c, command, host); err != nil {
			return fmt.Errorf("repl error: %w", err)
		}
		return nil
	}

	conn := rpc.NewConn(rpc.NewStreamTransport(host), &rpc.Options{
		BaseContext: func() context.Context { return c.Context },
		// BootstrapClient: export(),
	})
	defer conn.Close()

	client := conn.Bootstrap(c.Context)
	defer client.Release()

	f, release := system.Terminal(client).Login(c.Context, nil)
	defer release()

	res, err := f.Struct()
	if err != nil {
		return fmt.Errorf("failed to resolve terminal login: %w", err)
	}

	gs, err := NewGlobals(c)
	if err != nil {
		return err
	}

	eval := slurp.New()
	if err := eval.Bind(gs); err != nil {
		return fmt.Errorf("failed to bind base globals: %w", err)
	}
	// Bind session-specific globals (interactive mode only)
	if err := eval.Bind(NewSessionGlobals(c, &res)); err != nil {
		return fmt.Errorf("failed to bind session globals: %w", err)
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          c.String("prompt"),
		HistoryFile:     c.String("history-file"),
		AutoComplete:    getCompleter(c),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
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
		repl.WithReaderFactory(DefaultReaderFactory{IPFS: env.IPFS}),
		repl.WithInput(lineReader{Driver: rl}, func(err error) error {
			if err == nil || err == readline.ErrInterrupt {
				return nil
			}
			return err
		}),
	).Loop(c.Context); err != nil {
		return fmt.Errorf("repl: %w", err)
	}
	return nil
}

// executeCommand executes a single command line with a specific host file descriptor
func executeCommand(c *cli.Context, command string, host *os.File) error {
	conn := rpc.NewConn(rpc.NewStreamTransport(host), &rpc.Options{
		BaseContext: func() context.Context { return c.Context },
		// BootstrapClient: export(),
	})
	defer conn.Close()

	client := conn.Bootstrap(c.Context)
	defer client.Release()

	f, release := system.Terminal(client).Login(c.Context, nil)
	defer release()

	// Resolve the future to get the actual results
	res, err := f.Struct()
	if err != nil {
		return fmt.Errorf("failed to resolve terminal login: %w", err)
	}

	// Create base environment with analyzer and globals to it.
	eval := slurp.New()

	// Bind base globals (common to both modes)
	gs, err := NewGlobals(c)
	if err != nil {
		return fmt.Errorf("failed to bind base globals: %w", err)
	}
	if err := eval.Bind(gs); err != nil {
		return fmt.Errorf("failed to bind base globals: %w", err)
	}

	// Bind session-specific globals (including executor if --with-exec is set)
	if err := eval.Bind(NewSessionGlobals(c, &res)); err != nil {
		return fmt.Errorf("failed to bind session globals: %w", err)
	}

	// Create a reader from the command string
	commandReader := strings.NewReader(command)

	// Create a reader factory for IPFS path support
	readerFactory := DefaultReaderFactory{IPFS: env.IPFS}

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
	case path.Path:
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
func getCompleter(c *cli.Context) readline.AutoCompleter {
	completers := []readline.PrefixCompleterInterface{
		readline.PcItem("help"),
		readline.PcItem("version"),
		readline.PcItem("println"),
		readline.PcItem("print"),
		readline.PcItem("send"),
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
	}

	if c.Bool("with-ipfs") || c.Bool("with-all") {
		completers = append(completers, readline.PcItem("ipfs"))
	}
	if c.Bool("with-exec") || c.Bool("with-all") {
		completers = append(completers, readline.PcItem("exec"))
	}

	return readline.NewPrefixCompleter(completers...)
}

// NewSessionGlobals returns additional globals for interactive mode (requires terminal connection)
func NewSessionGlobals(c *cli.Context, f *system.Terminal_login_Results) map[string]core.Any {
	session := make(map[string]core.Any)

	// Add exec functionality if --with-exec flag is set
	if c.Bool("with-exec") || c.Bool("with-all") {
		session["exec"] = &Exec{Session: f}
	}

	return session
}
