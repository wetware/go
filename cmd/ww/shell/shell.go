package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/auth"
	"github.com/wetware/go/cmd/internal/flags"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/util"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "shell",
		Usage: "start an interactive REPL session",
		Description: `Start an interactive REPL session with controlled capabilities.

The shell can be run with different capability levels using the --with-* flags:
  --with-all           Grant all capabilities (console, IPFS, exec)
  --with-console       Grant console output capability only
  --with-ipfs          Grant IPFS capability only  
  --with-exec          Grant process execution capability only

If no capability flags are specified, the shell runs with minimal capabilities.
This provides a secure environment for testing and development.`,
		Flags: append([]cli.Flag{
			// &cli.StringSliceFlag{
			// 	Name:     "join",
			// 	Category: "P2P",
			// 	Aliases:  []string{"j"},
			// 	Usage:    "connect to cluster through specified peers",
			// 	EnvVars:  []string{"WW_JOIN"},
			// },
			// &cli.StringFlag{
			// 	Name:     "discover",
			// 	Category: "P2P",
			// 	Aliases:  []string{"d"},
			// 	Usage:    "automatic peer discovery settings",
			// 	Value:    "/mdns",
			// 	EnvVars:  []string{"WW_DISCOVER"},
			// },
			// &cli.StringFlag{
			// 	Name:     "namespace",
			// 	Category: "P2P",
			// 	Aliases:  []string{"ns"},
			// 	Usage:    "cluster namespace (must match dial host)",
			// 	Value:    "ww",
			// 	EnvVars:  []string{"WW_NAMESPACE"},
			// },
			// &cli.BoolFlag{
			// 	Name:     "dial",
			// 	Category: "P2P",
			// 	Usage:    "dial into a cluster using -join and -discover",
			// 	EnvVars:  []string{"WW_AUTODIAL"},
			// },
			// &cli.DurationFlag{
			// 	Name:     "timeout",
			// 	Category: "P2P",
			// 	Usage:    "timeout for -dial",
			// 	Value:    time.Second * 10,
			// },
		}, flags.CapabilityFlags()...),
		Action: func(c *cli.Context) error {
			// If no subcommand is specified, run the main action
			if c.Args().Len() == 0 {
				return Main(c)
			}

			// If a subcommand is specified, handle it
			subcommand := c.Args().First()
			if subcommand == "membrane" {
				ctx, cancel := context.WithCancel(c.Context)
				defer cancel()
				return util.DialSession(ctx, cell)
			}

			return cli.Exit("unknown subcommand: "+subcommand, 1)
		},
	}
}

func Main(c *cli.Context) error {
	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build the run command with capability flags
	args := []string{"run"}

	// Add capability flags if specified (these must come before the subcommand)
	if c.Bool("with-all") {
		args = append(args, "--with-all")
	}
	if c.Bool("with-console") {
		args = append(args, "--with-console")
	}
	if c.Bool("with-ipfs") {
		args = append(args, "--with-ipfs")
	}
	if c.Bool("with-exec") {
		args = append(args, "--with-exec")
	}

	// Add the subcommand and its arguments
	args = append(args, execPath, "shell", "membrane")

	cmd := exec.CommandContext(c.Context, execPath, args...)
	cmd.Stdin = c.App.Reader
	cmd.Stdout = c.App.Writer
	cmd.Stderr = c.App.ErrWriter

	cmd.Env = os.Environ()     // Inherit environment variables
	cmd.Dir = c.String("home") // Set the home directory for the command
	return cmd.Run()
}

func cell(ctx context.Context, sess auth.Terminal_login_Results) error {
	// Get user's home directory for history file
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to temp directory if home directory is not available
		home = os.TempDir()
	}

	env := core.New(lang.GlobalEnvConfig{
		Console:  sess.Console().AddRef(),
		IPFS:     sess.Ipfs().AddRef(),
		Executor: sess.Exec().AddRef(),
	}.New())

	interpreter := slurp.New(
		slurp.WithEnv(env),
		slurp.WithAnalyzer(lang.NewDotNotationAnalyzer(nil)))

	// Create readline input
	rlInput, err := NewReadlineInput(home)
	if err != nil {
		return fmt.Errorf("failed to create readline input: %w", err)
	}
	defer rlInput.Close()

	// Set the prompt on the readline instance
	rlInput.Prompt("ww » ")

	printer := printer{Writer: os.Stdout}

	// Custom REPL loop that handles errors gracefully
	for {
		// Read input
		input, err := rlInput.Readline()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// User pressed Ctrl+D, exit gracefully
				fmt.Println()
				break
			}
			// Other readline errors, continue
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}

		// Check for quit command
		if strings.TrimSpace(input) == "quit" {
			break
		}

		// Skip empty input
		if strings.TrimSpace(input) == "" {
			continue
		}

		// Create a reader for the input
		var rd *reader.Reader
		if sess.HasIpfs() {
			ipfs := sess.Ipfs()
			rd = readerFactory(ipfs)(strings.NewReader(input))
		} else {
			rd = lang.NewReaderWithHexSupport(strings.NewReader(input))
		}

		// Read and evaluate
		form, err := rd.One()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		// Evaluate the form
		result, err := interpreter.Eval(form)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		if err := printer.Print(result); err != nil {
			fmt.Fprintf(os.Stderr, "Error printing result: %v\n", err)
		}
	}

	return nil
}
