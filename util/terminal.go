package util

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Terminal formatting utilities for enhanced output

// ANSI color codes
const (
	ColorReset     = "\033[0m"
	ColorRed       = "\033[31m"
	ColorGreen     = "\033[32m"
	ColorYellow    = "\033[33m"
	ColorBlue      = "\033[34m"
	ColorMagenta   = "\033[35m"
	ColorCyan      = "\033[36m"
	ColorWhite     = "\033[37m"
	ColorBold      = "\033[1m"
	ColorDim       = "\033[2m"
	ColorItalic    = "\033[3m"
	ColorUnderline = "\033[4m"
)

// TerminalInfo holds information about the terminal capabilities
type TerminalInfo struct {
	IsTTY         bool
	Width         int
	Height        int
	SupportsColor bool
}

// GetTerminalInfo returns information about the current terminal
func GetTerminalInfo() *TerminalInfo {
	info := &TerminalInfo{
		IsTTY:         isTTY(),
		SupportsColor: supportsColor(),
	}

	if info.IsTTY {
		info.Width, info.Height = getTerminalSize()
	}

	return info
}

// isTTY checks if stdout is a terminal
func isTTY() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// supportsColor checks if the terminal supports colors
func supportsColor() bool {
	// Check for common environment variables that indicate color support
	colorTerm := os.Getenv("COLORTERM")
	term := os.Getenv("TERM")

	if colorTerm != "" {
		return true
	}

	if strings.Contains(term, "color") || strings.Contains(term, "xterm") {
		return true
	}

	return false
}

// getTerminalSize returns the terminal dimensions
func getTerminalSize() (width, height int) {
	// Default values
	width, height = 80, 24

	// Try to get actual terminal size
	// This is a simplified implementation
	// In a real implementation, you might use syscall or a library like "golang.org/x/term"

	return width, height
}

// FormatTable formats data as a table with proper alignment
func FormatTable(headers []string, rows [][]string, useColors bool) string {
	if len(rows) == 0 {
		return ""
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var result strings.Builder

	// Header
	if useColors {
		result.WriteString(ColorBold)
	}
	result.WriteString(formatRow(headers, colWidths))
	if useColors {
		result.WriteString(ColorReset)
	}
	result.WriteString("\n")

	// Separator
	separator := make([]string, len(headers))
	for i := range headers {
		separator[i] = strings.Repeat("-", colWidths[i])
	}
	if useColors {
		result.WriteString(ColorDim)
	}
	result.WriteString(formatRow(separator, colWidths))
	if useColors {
		result.WriteString(ColorReset)
	}
	result.WriteString("\n")

	// Rows
	for _, row := range rows {
		result.WriteString(formatRow(row, colWidths))
		result.WriteString("\n")
	}

	return result.String()
}

// formatRow formats a single row with proper column alignment
func formatRow(row []string, colWidths []int) string {
	var result strings.Builder

	for i, cell := range row {
		if i >= len(colWidths) {
			break
		}

		// Pad the cell to the column width
		padded := fmt.Sprintf("%-*s", colWidths[i], cell)
		result.WriteString(padded)

		// Add spacing between columns
		if i < len(row)-1 {
			result.WriteString("  ")
		}
	}

	return result.String()
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}

	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// FormatBytes formats bytes in a human-readable way
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ProgressBar creates a simple progress bar
func ProgressBar(current, total int, width int, useColors bool) string {
	if total <= 0 {
		return ""
	}

	percentage := float64(current) / float64(total)
	filled := int(float64(width) * percentage)

	var bar strings.Builder
	bar.WriteString("[")

	for i := 0; i < width; i++ {
		if i < filled {
			if useColors {
				bar.WriteString(ColorGreen + "=" + ColorReset)
			} else {
				bar.WriteString("=")
			}
		} else {
			bar.WriteString(" ")
		}
	}

	bar.WriteString("]")
	bar.WriteString(fmt.Sprintf(" %d%%", int(percentage*100)))

	return bar.String()
}

// Spinner provides a simple loading spinner
type Spinner struct {
	frames []string
	index  int
}

// NewSpinner creates a new spinner with default frames
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		index:  0,
	}
}

// Next returns the next spinner frame
func (s *Spinner) Next() string {
	frame := s.frames[s.index]
	s.index = (s.index + 1) % len(s.frames)
	return frame
}

// StatusIndicator provides status indicators for the prompt
type StatusIndicator struct {
	spinner *Spinner
	start   time.Time
}

// NewStatusIndicator creates a new status indicator
func NewStatusIndicator() *StatusIndicator {
	return &StatusIndicator{
		spinner: NewSpinner(),
		start:   time.Now(),
	}
}

// GetStatus returns a formatted status string
func (s *StatusIndicator) GetStatus(useColors bool) string {
	duration := time.Since(s.start)

	if useColors {
		return fmt.Sprintf("%s%s%s %s",
			ColorYellow, s.spinner.Next(), ColorReset,
			ColorDim+FormatDuration(duration)+ColorReset)
	}

	return fmt.Sprintf("%s %s", s.spinner.Next(), FormatDuration(duration))
}
