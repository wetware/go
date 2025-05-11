package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"strings"
	"testing"
)

// TestEcho verifies the core echo functionality with various input types.
// It tests the _echo function directly with different inputs including:
// - Empty input
// - Single line text
// - Multiple lines
// - Binary data
func TestEcho(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "hello world\n",
			expected: "hello world\n",
		},
		{
			name:     "multiple lines",
			input:    "line 1\nline 2\nline 3\n",
			expected: "line 1\nline 2\nline 3\n",
		},
		{
			name:     "binary data",
			input:    string([]byte{0x00, 0x01, 0x02, 0x03}),
			expected: string([]byte{0x00, 0x01, 0x02, 0x03}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := _echo(&buf, strings.NewReader(tt.input))
			if err != nil {
				t.Errorf("_echo() error = %v", err)
				return
			}
			if got := buf.String(); got != tt.expected {
				t.Errorf("_echo() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestEchoNilInput verifies that the echo function properly handles nil input.
func TestEchoNilInput(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := _echo(&buf, nil)
	if err == nil {
		t.Error("_echo() with nil reader should return error")
	}
}

// TestEchoNilOutput verifies that the echo function properly handles nil output.
func TestEchoNilOutput(t *testing.T) {
	t.Parallel()
	err := _echo(nil, strings.NewReader("test"))
	if err == nil {
		t.Error("_echo() with nil writer should return error")
	}
}

// TestOperationalModes verifies the program's different modes of operation:
//  1. Stdin mode: reads from standard input and writes to standard output
//     Example: echo "hello world" | ./echo -stdin
//  2. Args mode: takes command-line arguments and writes them to standard output
//     Example: ./echo arg1 arg2 arg3
func TestOperationalModes(t *testing.T) {
	t.Parallel()
	t.Run("stdin mode", func(t *testing.T) {
		// Save original stdin and stdout
		oldStdin := os.Stdin
		oldStdout := os.Stdout
		defer func() {
			os.Stdin = oldStdin
			os.Stdout = oldStdout
		}()

		// Create test input
		input := "test input\n"
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = r

		// Create test output pipe
		outR, outW, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = outW

		// Write test input
		go func() {
			w.WriteString(input)
			w.Close()
		}()

		// Reset flag state
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		// Set stdin flag and run main
		os.Args = []string{"echo", "-stdin"}
		main()

		// Close the write end of the output pipe
		outW.Close()

		// Read the output
		var buf bytes.Buffer
		io.Copy(&buf, outR)

		// Check output
		if got := buf.String(); got != input {
			t.Errorf("main() with stdin = %v, want %v", got, input)
		}
	})

	t.Run("args mode", func(t *testing.T) {
		// Save original stdout
		oldStdout := os.Stdout
		defer func() {
			os.Stdout = oldStdout
		}()

		// Create test output pipe
		outR, outW, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = outW

		// Reset flag state
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		// Set test args
		os.Args = []string{"echo", "arg1", "arg2", "arg3"}
		main()

		// Close the write end of the output pipe
		outW.Close()

		// Read the output
		var buf bytes.Buffer
		io.Copy(&buf, outR)

		// Check output
		expected := "arg1 arg2 arg3 "
		if got := buf.String(); got != expected {
			t.Errorf("main() with args = %v, want %v", got, expected)
		}
	})
}
