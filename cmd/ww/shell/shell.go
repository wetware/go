package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/chzyer/readline"
	"github.com/spy16/slurp"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
	"github.com/spy16/slurp/repl"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/auth"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/util"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:   "shell",
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
	cmd.Env = os.Environ() // Inherit environment variables
	return cmd.Run()
}

// ReadlineInput implements the repl.Input interface using github.com/chzyer/readline
type ReadlineInput struct {
	rl *readline.Instance
}

// NewReadlineInput creates a new readline-based input
func NewReadlineInput(prompt string) (*ReadlineInput, error) {
	rl, err := readline.New(prompt)
	if err != nil {
		return nil, err
	}
	return &ReadlineInput{rl: rl}, nil
}

// Readline implements repl.Input.Readline
func (r *ReadlineInput) Readline() (string, error) {
	return r.rl.Readline()
}

// Prompt implements repl.Prompter.Prompt
func (r *ReadlineInput) Prompt(prompt string) {
	r.rl.SetPrompt(prompt)
}

// Close closes the readline instance
func (r *ReadlineInput) Close() error {
	return r.rl.Close()
}

func cell(ctx context.Context, sess auth.Terminal_login_Results) error {
	ipfs := sess.Ipfs()
	defer ipfs.Release()

	ipfsWrapper := lang.Session{IPFS: ipfs}
	env := core.New(map[string]core.Any{
		"ipfs": ipfsWrapper,
	})

	interpreter := slurp.New(
		slurp.WithEnv(env),
		slurp.WithAnalyzer(nil))

	// Create a REPL with multiline support and readline input
	repl := repl.New(interpreter,
		repl.WithBanner("Wetware Shell - Type 'quit' to exit"),
		// repl.WithInput()
		// repl.WithPrinter()
		repl.WithPrompts("ww ", "  | "),
		repl.WithReaderFactory(readerFactory()),
	)

	// Start the REPL loop
	return repl.Loop(ctx)
}

func readerFactory() repl.ReaderFactoryFunc {
	return func(r io.Reader) *reader.Reader {
		rd := reader.New(r)

		// Set up the Unix path reader macro for '/' character
		rd.SetMacro('/', false, lang.UnixPathReader())

		return rd
	}
}
