package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

// ReadlineInput implements the repl.Input interface using github.com/chzyer/readline
type ReadlineInput struct {
	rl *readline.Instance
}

// NewReadlineInput creates a new readline-based input with enhanced configuration
func NewReadlineInput(home string) (*ReadlineInput, error) {
	// Create history file path in the user's home directory
	historyFile := filepath.Join(home, ".ww_history")

	// Ensure the history file directory exists and is writable
	if err := os.MkdirAll(filepath.Dir(historyFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	// Enhanced readline configuration with better formatting and features
	rl, err := readline.NewEx(&readline.Config{
		HistoryFile: historyFile,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,

		// Enhanced prompts with colors and status information
		InterruptPrompt: "\033[33m⏎\033[0m",    // Yellow interrupt symbol
		EOFPrompt:       "\033[31mexit\033[0m", // Red exit prompt

		// Enable advanced features
		DisableAutoSaveHistory: false,
		HistorySearchFold:      true,                // Case-insensitive history search
		HistoryLimit:           10000,               // Increase history limit to 10000 entries
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
	// Ensure history is saved before closing
	if err := r.rl.SaveHistory(""); err != nil {
		// Log the error but don't fail the close operation
		fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
	}
	return r.rl.Close()
}
