package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

/*
Supported Keyboard Shortcuts:
- Ctrl+A: Move to beginning of line
- Ctrl+E: Move to end of line
- Ctrl+K: Kill line from cursor to end
- Ctrl+U: Kill line from beginning to cursor
- Ctrl+W: Kill word before cursor
- Ctrl+Y: Yank (paste) previously killed text
- Ctrl+T: Transpose characters
- Ctrl+L: Clear screen
- Ctrl+R: Reverse history search
- Arrow keys: Navigate through line and history
- Tab: Auto-completion
- Backspace/Delete: Normal text editing
*/

// InteractiveInput implements the repl.Input interface using github.com/chzyer/readline
type InteractiveInput struct {
	*readline.Instance
}

// NewReadlineInput creates a new readline-based input with enhanced configuration
func NewReadlineInput(home string) (*InteractiveInput, error) {
	// Create history file path in the user's home directory
	historyFile := filepath.Join(home, ".ww_history")

	// Ensure the history file directory exists and is writable
	if err := os.MkdirAll(filepath.Dir(historyFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	// Enhanced readline configuration with better formatting and features
	i, err := readline.NewEx(&readline.Config{
		HistoryFile: historyFile,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Stdin:       os.Stdin, // Explicitly set stdin for proper input handling

		// Enhanced prompts with colors and status information
		InterruptPrompt: "\033[33m⏎\033[0m",    // Yellow interrupt symbol
		EOFPrompt:       "\033[31mexit\033[0m", // Red exit prompt

		// Enable advanced features
		DisableAutoSaveHistory: false,
		HistorySearchFold:      true,                // Case-insensitive history search
		HistoryLimit:           10000,               // Increase history limit to 10k entries
		AutoComplete:           &WetwareCompleter{}, // Custom autocomplete

		// Enhanced display settings
		UniqueEditLine: true,
		Listener:       &WetwareListener{}, // Custom listener for status updates

		// Add terminal width function for proper line wrapping
		FuncGetWidth: func() int {
			// Get terminal width for proper line wrapping
			return 80 // Default fallback
		},
	})
	if err != nil {
		return nil, err
	}
	return &InteractiveInput{Instance: i}, nil
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

// Enhanced WetwareListener with better keyboard shortcut support
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
	case 1: // Ctrl+A - Move to beginning of line
		return line, 0, true
	case 5: // Ctrl+E - Move to end of line
		return line, len(line), true
	case 11: // Ctrl+K - Kill line from cursor to end
		return line[:pos], pos, true
	case 21: // Ctrl+U - Kill line from beginning to cursor
		return line[pos:], 0, true
	case 23: // Ctrl+W - Kill word before cursor
		return l.killWordBefore(line, pos), l.wordStart(line, pos), true
	case 25: // Ctrl+Y - Yank (paste) previously killed text
		return line, pos, true
	case 20: // Ctrl+T - Transpose characters
		if pos > 0 && pos < len(line) {
			newLine := make([]rune, len(line))
			copy(newLine, line)
			newLine[pos-1], newLine[pos] = newLine[pos], newLine[pos-1]
			return newLine, pos, true
		}
	case 27: // ESC - Handle arrow keys and other escape sequences
		// This will be handled by the readline library's built-in arrow key support
		return line, pos, false
	}
	return line, pos, false
}

// Helper methods for word manipulation
func (l *WetwareListener) killWordBefore(line []rune, pos int) []rune {
	if pos == 0 {
		return line
	}
	start := l.wordStart(line, pos)
	return append(line[:start], line[pos:]...)
}

func (l *WetwareListener) wordStart(line []rune, pos int) int {
	if pos == 0 {
		return 0
	}
	start := pos - 1
	for start >= 0 && !isWordBoundary(line[start]) {
		start--
	}
	return start + 1
}

// Readline implements repl.Input.Readline
func (r *InteractiveInput) Readline() (string, error) {
	for {
		switch line, err := r.Instance.Readline(); err {
		case readline.ErrInterrupt:
			if len(line) == 0 {
				// Enhanced interrupt handling with visual feedback
				r.Instance.Clean()
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
func (r *InteractiveInput) Prompt(prompt string) {
	// Enhanced prompt with colors and status information
	enhancedPrompt := r.enhancePrompt(prompt)
	r.Instance.SetPrompt(enhancedPrompt)
}

// enhancePrompt adds colors and status information to the prompt
func (r *InteractiveInput) enhancePrompt(basePrompt string) string {
	// Add colors and status indicators
	status := r.getStatusInfo()

	// Format: ww >> [status] prompt
	return fmt.Sprintf("\033[36m%s\033[0m %s \033[32m%s\033[0m ",
		basePrompt, status, "›")
}

// getStatusInfo returns status information for the prompt
func (r *InteractiveInput) getStatusInfo() string {
	// This could be enhanced to show actual system status
	// For now, return a simple indicator
	return "\033[33m●\033[0m" // Yellow dot indicator
}

// Close closes the readline instance
func (r *InteractiveInput) Close() error {
	// Ensure history is saved before closing
	if err := r.Instance.SaveHistory(""); err != nil {
		// Log the error but don't fail the close operation
		fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
	}
	return r.Instance.Close()
}
