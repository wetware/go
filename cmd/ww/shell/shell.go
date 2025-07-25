package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/repl"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/auth"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/util"
)

var (
	flags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "join",
			Aliases: []string{"j"},
			Usage:   "connect to cluster through specified peers",
			EnvVars: []string{"WW_JOIN"},
		},
		&cli.StringFlag{
			Name:    "discover",
			Aliases: []string{"d"},
			Usage:   "automatic peer discovery settings",
			Value:   "/mdns",
			EnvVars: []string{"WW_DISCOVER"},
		},
		&cli.StringFlag{
			Name:    "namespace",
			Aliases: []string{"ns"},
			Usage:   "cluster namespace (must match dial host)",
			Value:   "ww",
			EnvVars: []string{"WW_NAMESPACE"},
		},
		&cli.BoolFlag{
			Name:    "quiet",
			Aliases: []string{"q"},
			Usage:   "suppress banner message on interactive startup",
			EnvVars: []string{"WW_QUIET"},
		},
		&cli.BoolFlag{
			Name:    "dial",
			Usage:   "dial into a cluster using -join and -discover",
			EnvVars: []string{"WW_AUTODIAL"},
		},
		&cli.DurationFlag{
			Name:  "timeout",
			Usage: "timeout for -dial",
			Value: time.Second * 10,
		},
		// Capability control flags (same as ww run)
		&cli.BoolFlag{
			Name:    "with-full-rights",
			Usage:   "grant all capabilities (console, IPFS, exec)",
			EnvVars: []string{"WW_WITH_FULL_RIGHTS"},
		},
		&cli.BoolFlag{
			Name:    "with-console",
			Usage:   "grant console output capability",
			EnvVars: []string{"WW_WITH_CONSOLE"},
		},
		&cli.BoolFlag{
			Name:    "with-ipfs",
			Usage:   "grant IPFS capability",
			EnvVars: []string{"WW_WITH_IPFS"},
		},
		&cli.BoolFlag{
			Name:    "with-exec",
			Usage:   "grant process execution capability",
			EnvVars: []string{"WW_WITH_EXEC"},
		},
	}
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "shell",
		Usage: "start an interactive REPL session",
		Description: `Start an interactive REPL session with controlled capabilities.

The shell can be run with different capability levels using the --with-* flags:
  --with-full-rights    Grant all capabilities (console, IPFS, exec)
  --with-console        Grant console output capability only
  --with-ipfs           Grant IPFS capability only  
  --with-exec           Grant process execution capability only

If no capability flags are specified, the shell runs with minimal capabilities.
This provides a secure environment for testing and development.`,
		Flags:  flags,
		Action: Main,
		Subcommands: []*cli.Command{
			{
				Name:   "membrane",
				Hidden: true,
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()

					return util.DialSession(ctx, cell)
				},
			},
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
	args := []string{"run", execPath, "shell", "membrane"}

	// Add capability flags if specified
	if c.Bool("with-full-rights") {
		args = append(args, "--with-full-rights")
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
	envMap := make(map[string]core.Any)

	// Conditionally add console capability
	if sess.HasConsole() {
		console := sess.Console()
		defer console.Release()
		envMap["println"] = lang.ConsolePrintln{Console: console}
	}

	// Conditionally add IPFS capabilities
	if sess.HasIpfs() {
		ipfs := sess.Ipfs()
		defer ipfs.Release()

		envMap["cat"] = lang.IPFSCat{IPFS: ipfs}
		envMap["add"] = lang.IPFSAdd{IPFS: ipfs}
		envMap["ls"] = &lang.IPFSLs{IPFS: ipfs}
		envMap["stat"] = &lang.IPFSStat{IPFS: ipfs}
		envMap["pin"] = &lang.IPFSPin{IPFS: ipfs}
		envMap["unpin"] = &lang.IPFSUnpin{IPFS: ipfs}
		envMap["pins"] = &lang.IPFSPins{IPFS: ipfs}
		envMap["id"] = &lang.IPFSId{IPFS: ipfs}
		envMap["connect"] = &lang.IPFSConnect{IPFS: ipfs}
		envMap["peers"] = &lang.IPFSPeers{IPFS: ipfs}
	}

	// Conditionally add process execution capability
	if sess.HasExec() {
		exec := sess.Exec()
		defer exec.Release()
		envMap["go"] = lang.Go{Executor: exec}
	}

	env := core.New(envMap)

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
