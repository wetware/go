package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/repl"
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
			&cli.BoolFlag{
				Name:     "quiet",
				Category: "OUTPUT",
				Aliases:  []string{"q"},
				Usage:    "suppress banner message on interactive startup",
				EnvVars:  []string{"WW_QUIET"},
			},
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

	// Initialize environment with basic functions
	globals := make(map[string]core.Any)

	// Conditionally add console capability
	if sess.HasConsole() {
		console := sess.Console()
		defer console.Release()
		globals["println"] = lang.ConsolePrintln{Console: console}
	}

	// Conditionally add IPFS capabilities
	if sess.HasIpfs() {
		ipfs := sess.Ipfs()
		defer ipfs.Release()

		globals["cat"] = lang.IPFSCat{IPFS: ipfs}
		globals["add"] = lang.IPFSAdd{IPFS: ipfs}
		globals["ls"] = &lang.IPFSLs{IPFS: ipfs}
		globals["stat"] = &lang.IPFSStat{IPFS: ipfs}
		globals["pin"] = &lang.IPFSPin{IPFS: ipfs}
		globals["unpin"] = &lang.IPFSUnpin{IPFS: ipfs}
		globals["pins"] = &lang.IPFSPins{IPFS: ipfs}
		globals["id"] = &lang.IPFSId{IPFS: ipfs}
		globals["connect"] = &lang.IPFSConnect{IPFS: ipfs}
		globals["peers"] = &lang.IPFSPeers{IPFS: ipfs}
	}

	// Conditionally add process execution capability
	if sess.HasExec() {
		exec := sess.Exec()
		defer exec.Release()
		globals["go"] = lang.Go{Executor: exec}
	}

	env := core.New(globals)

	interpreter := slurp.New(
		slurp.WithEnv(env),
		slurp.WithAnalyzer(nil))

	// Create readline input
	rlInput, err := NewReadlineInput(home)
	if err != nil {
		return fmt.Errorf("failed to create readline input: %w", err)
	}
	defer rlInput.Close()

	// Create a REPL with readline input
	replConfig := []repl.Option{
		repl.WithBanner("Wetware Shell - Type 'quit' to exit"),
		repl.WithPrompts("ww »", "   ›"),
		repl.WithPrinter(printer{}),
		repl.WithInput(rlInput, nil),
	}

	// Conditionally add IPFS reader factory if IPFS is available
	if sess.HasIpfs() {
		ipfs := sess.Ipfs()
		replConfig = append(replConfig, repl.WithReaderFactory(readerFactory(ipfs)))
	}

	repl := repl.New(interpreter, replConfig...)

	// Start the REPL loop
	return repl.Loop(ctx)
}
