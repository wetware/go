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
	}
)

func Command() *cli.Command {
	return &cli.Command{
		Name:   "shell",
		Usage:  "start an interactive REPL session",
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

	cmd := exec.CommandContext(c.Context, execPath, "run", execPath, "shell", "membrane")
	cmd.Stdin = c.App.Reader
	cmd.Stdout = c.App.Writer
	cmd.Stderr = c.App.ErrWriter

	cmd.Env = os.Environ()     // Inherit environment variables
	cmd.Dir = c.String("home") // Set the home directory for the command
	return cmd.Run()
}

func cell(ctx context.Context, sess auth.Terminal_login_Results) error {
	ipfs := sess.Ipfs()
	defer ipfs.Release()

	exec := sess.Exec()
	defer exec.Release()

	console := sess.Console()
	defer console.Release()

	// Get user's home directory for history file
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to temp directory if home directory is not available
		home = os.TempDir()
	}

	env := core.New(map[string]core.Any{
		// Console
		"println": lang.ConsolePrintln{Console: console},

		// IPFS
		"cat":     lang.IPFSCat{IPFS: ipfs},
		"add":     lang.IPFSAdd{IPFS: ipfs},
		"ls":      &lang.IPFSLs{IPFS: ipfs},
		"stat":    &lang.IPFSStat{IPFS: ipfs},
		"pin":     &lang.IPFSPin{IPFS: ipfs},
		"unpin":   &lang.IPFSUnpin{IPFS: ipfs},
		"pins":    &lang.IPFSPins{IPFS: ipfs},
		"id":      &lang.IPFSId{IPFS: ipfs},
		"connect": &lang.IPFSConnect{IPFS: ipfs},
		"peers":   &lang.IPFSPeers{IPFS: ipfs},

		// Process execution
		"go": lang.Go{Executor: exec},
	})

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
	repl := repl.New(interpreter,
		repl.WithBanner("Wetware Shell - Type 'quit' to exit"),
		repl.WithPrompts("ww »", "   ›"),
		repl.WithReaderFactory(readerFactory(ipfs)),
		repl.WithPrinter(printer{}),
		repl.WithInput(rlInput, nil),
	)

	// Start the REPL loop
	return repl.Loop(ctx)
}
