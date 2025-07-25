package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/chzyer/readline"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
	"github.com/spy16/slurp/repl"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/auth"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
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

// ReadlineInput implements the repl.Input interface using github.com/chzyer/readline
type ReadlineInput struct {
	rl *readline.Instance
}

// NewReadlineInput creates a new readline-based input with enhanced configuration
func NewReadlineInput(home string) (*ReadlineInput, error) {
	// Enhanced readline configuration with better formatting and features
	rl, err := readline.NewEx(&readline.Config{
		HistoryFile: "history.ww",
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,

		// Enhanced prompts with colors and status information
		InterruptPrompt: "\033[33m⏎\033[0m",    // Yellow interrupt symbol
		EOFPrompt:       "\033[31mexit\033[0m", // Red exit prompt

		// Enable advanced features
		DisableAutoSaveHistory: false,
		HistorySearchFold:      true,                // Case-insensitive history search
		AutoComplete:           &WetwareCompleter{}, // Custom autocomplete

		// Enhanced display settings
		UniqueEditLine: true,
		Listener:       &WetwareListener{}, // Custom listener for status updates
	})
	if err != nil {
		return nil, err
	}
	return &ReadlineInput{rl: rl}, nil
}

// WetwareCompleter provides intelligent autocomplete for the shell
type WetwareCompleter struct{}

func (c *WetwareCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	// Get the current word being typed
	word := getCurrentWord(line, pos)
	if len(word) == 0 {
		return nil, 0
	}

	// Built-in commands and functions
	commands := []string{
		"cat", "add", "ls", "stat", "pin", "unpin", "pins", "id", "connect", "peers", "go",
		"quit", "exit", "help", "clear", "history", "cd", "pwd",
	}

	var suggestions [][]rune
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, word) {
			suggestions = append(suggestions, []rune(cmd))
		}
	}

	// If we have suggestions, return them
	if len(suggestions) > 0 {
		return suggestions, len(word)
	}

	return nil, 0
}

// getCurrentWord extracts the current word being typed
func getCurrentWord(line []rune, pos int) string {
	if pos == 0 {
		return ""
	}

	start := pos - 1
	for start >= 0 && !isWordBoundary(line[start]) {
		start--
	}
	start++

	return string(line[start:pos])
}

// isWordBoundary checks if a character is a word boundary
func isWordBoundary(r rune) bool {
	return r == ' ' || r == '\t' || r == '(' || r == ')' || r == '[' || r == ']'
}

// WetwareListener provides custom event handling for the readline instance
type WetwareListener struct{}

func (l *WetwareListener) OnChange(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
	// Handle special key combinations
	switch key {
	case readline.CharCtrlL:
		// Clear screen
		fmt.Print("\033[2J\033[H")
		return line, pos, true
	case 18: // Ctrl+R
		// Trigger history search
		return line, pos, true
	}
	return line, pos, false
}

// Readline implements repl.Input.Readline
func (r *ReadlineInput) Readline() (string, error) {
	for {
		switch line, err := r.rl.Readline(); err {
		case readline.ErrInterrupt:
			if len(line) == 0 {
				// Enhanced interrupt handling with visual feedback
				r.rl.Clean()
				fmt.Fprintf(os.Stderr, "\033[33m^C\033[0m\n") // Yellow ^C indicator
				return "", nil
			}
			continue
		default:
			return line, err // io.EOF or other errors
		}
	}
}

// Prompt implements repl.Prompter.Prompt with enhanced formatting
func (r *ReadlineInput) Prompt(prompt string) {
	// Enhanced prompt with colors and status information
	enhancedPrompt := r.enhancePrompt(prompt)
	r.rl.SetPrompt(enhancedPrompt)
}

// enhancePrompt adds colors and status information to the prompt
func (r *ReadlineInput) enhancePrompt(basePrompt string) string {
	// Add colors and status indicators
	status := r.getStatusInfo()

	// Format: ww >> [status] prompt
	return fmt.Sprintf("\033[36m%s\033[0m %s \033[32m%s\033[0m ",
		basePrompt, status, "›")
}

// getStatusInfo returns status information for the prompt
func (r *ReadlineInput) getStatusInfo() string {
	// This could be enhanced to show actual system status
	// For now, return a simple indicator
	return "\033[33m●\033[0m" // Yellow dot indicator
}

// Close closes the readline instance
func (r *ReadlineInput) Close() error {
	return r.rl.Close()
}

func cell(ctx context.Context, sess auth.Terminal_login_Results) error {
	ipfs := sess.Ipfs()
	defer ipfs.Release()

	exec := sess.Exec()
	defer exec.Release()

	// Get user's home directory for history file
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to temp directory if home directory is not available
		home = os.TempDir()
	}

	env := core.New(map[string]core.Any{
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

func readerFactory(ipfs system.IPFS) repl.ReaderFactoryFunc {
	return func(r io.Reader) *reader.Reader {
		// Create a reader with hex support
		rd := lang.NewReaderWithHexSupport(r)

		// Set up the Unix path reader macro for '/' character
		rd.SetMacro('/', false, lang.UnixPathReader())
		// Set up the custom list reader macro for '(' character
		rd.SetMacro('(', false, lang.ListReader(ipfs))

		return rd
	}
}

type printer struct{}

// ANSI color codes for enhanced output formatting
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
)

func (printer) Print(val interface{}) error {
	if err, ok := val.(error); ok {
		// Enhanced error formatting with colors
		fmt.Fprintf(os.Stdout, "%s%s%s\n", colorRed, err.Error(), colorReset)
		return nil
	}

	// Handle different types for rendering with enhanced formatting
	switch v := val.(type) {

	case *lang.Buffer:
		// Enhanced buffer display with hex preview
		if len(v.Mem) > 0 {
			fmt.Fprintf(os.Stdout, "%sBuffer (%d bytes):%s\n", colorBold, len(v.Mem), colorReset)
			fmt.Fprintf(os.Stdout, "%s%s%s\n", colorCyan, v.String(), colorReset)
			if len(v.Mem) <= 64 {
				fmt.Fprintf(os.Stdout, "%sHex: %s%s\n", colorDim, v.AsHex(), colorReset)
			}
		} else {
			fmt.Fprintf(os.Stdout, "%sEmpty buffer%s\n", colorYellow, colorReset)
		}

	case core.SExpressable:
		form, err := v.SExpr()
		if err != nil {
			return err
		}
		// Enhanced s-expression formatting
		fmt.Fprintf(os.Stdout, "%s%s%s\n", colorDim, form, colorReset)

	case string:
		// Enhanced string output with syntax highlighting for IPFS paths
		if strings.HasPrefix(v, "/ipfs/") || strings.HasPrefix(v, "/ipld/") {
			fmt.Fprintf(os.Stdout, "%s%s%s\n", colorBlue, v, colorReset)
		} else {
			fmt.Fprintf(os.Stdout, "%s%s%s\n", colorGreen, v, colorReset)
		}

	case lang.Map:
		// Pretty print maps with indentation
		printMap(v, 0, true)

	case core.Any:
		// For core.Any types, try to convert to string
		if str, ok := v.(string); ok {
			fmt.Fprintf(os.Stdout, "%s%s%s\n", colorGreen, str, colorReset)
		} else {
			fmt.Fprintf(os.Stdout, "%s%+v%s\n", colorYellow, v, colorReset)
		}
	default:
		fmt.Fprintf(os.Stdout, "%s%+v%s\n", colorYellow, v, colorReset)
	}
	return nil
}

// printMap recursively prints a map with proper indentation and colors
func printMap(m lang.Map, indent int, useColors bool) {
	indentStr := strings.Repeat("  ", indent)

	for key, value := range m {
		keyStr := fmt.Sprintf("%v", key)

		// Print key with color
		if useColors {
			fmt.Fprintf(os.Stdout, "%s%s%s%s: ", indentStr, colorCyan, keyStr, colorReset)
		} else {
			fmt.Fprintf(os.Stdout, "%s%s: ", indentStr, keyStr)
		}

		// Handle nested maps recursively
		if nestedMap, ok := value.(lang.Map); ok {
			fmt.Fprintf(os.Stdout, "\n")
			printMap(nestedMap, indent+1, useColors)
		} else {
			// Print value with appropriate color
			if useColors {
				fmt.Fprintf(os.Stdout, "%s%v%s\n", colorGreen, value, colorReset)
			} else {
				fmt.Fprintf(os.Stdout, "%v\n", value)
			}
		}
	}
}
