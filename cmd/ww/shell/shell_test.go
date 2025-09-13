package shell

import (
	"bytes"
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

// createMockCLIContext creates a mock CLI context for testing
func createMockCLIContext() *cli.Context {
	app := &cli.App{}
	app.Flags = []cli.Flag{}
	flagSet := &flag.FlagSet{}
	flagSet.Bool("with-ipfs", false, "Enable IPFS capability") // Set to false for tests
	flagSet.Bool("with-exec", true, "Enable exec capability")
	flagSet.Bool("with-console", true, "Enable console capability")
	flagSet.Bool("with-all", false, "Enable all capabilities")
	flagSet.String("prompt", "ww> ", "Shell prompt")
	flagSet.String("history-file", "/tmp/ww_history", "History file path")
	ctx := cli.NewContext(app, flagSet, nil)
	ctx.Context = context.Background()
	return ctx
}

func TestExecuteCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		command   string
		wantError bool
	}{
		{
			name:      "simple arithmetic",
			command:   "(+ 1 2 3)",
			wantError: false,
		},
		{
			name:      "multiplication",
			command:   "(* 2 3 4)",
			wantError: false,
		},
		{
			name:      "comparison",
			command:   "(> 5 3)",
			wantError: false,
		},
		{
			name:      "equality",
			command:   "(= 5 5)",
			wantError: false,
		},
		{
			name:      "boolean values",
			command:   "true",
			wantError: false,
		},
		{
			name:      "nil value",
			command:   "nil",
			wantError: false,
		},
		{
			name:      "version",
			command:   "version",
			wantError: false,
		},
		{
			name:      "help function",
			command:   "(help)",
			wantError: false,
		},
		{
			name:      "println function",
			command:   "(println \"Hello World\")",
			wantError: false,
		},
		{
			name:      "nested expressions",
			command:   "(+ (* 2 3) 4)",
			wantError: false,
		},
		{
			name:      "invalid syntax",
			command:   "(+ 1 2",
			wantError: true,
		},
		{
			name:      "unknown function",
			command:   "(unknown-function 1 2)",
			wantError: true,
		},
		{
			name:      "empty command",
			command:   "",
			wantError: false,
		},
		{
			name:      "whitespace only",
			command:   "   ",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeCommand(createMockCLIContext(), tt.command)

			if tt.wantError {
				assert.Error(t, err, "Expected error for command: %s", tt.command)
			} else {
				assert.NoError(t, err, "Expected no error for command: %s", tt.command)
			}
		})
	}
}

func TestExecuteCommandWithIPFS(t *testing.T) {
	t.Parallel()

	// These tests will fail if IPFS is not available, but they test the structure
	tests := []struct {
		name      string
		command   string
		wantError bool
	}{
		{
			name:      "ipfs function exists",
			command:   "ipfs",
			wantError: true, // IPFS not available in test environment
		},
		{
			name:      "ipfs cat with invalid path",
			command:   "(ipfs :cat \"/invalid/path\")",
			wantError: true, // Should fail with invalid path
		},
		{
			name:      "ipfs get with invalid path",
			command:   "(ipfs :get \"/invalid/path\")",
			wantError: true, // Should fail with invalid path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeCommand(createMockCLIContext(), tt.command)

			if tt.wantError {
				assert.Error(t, err, "Expected error for command: %s", tt.command)
			} else {
				assert.NoError(t, err, "Expected no error for command: %s", tt.command)
			}
		})
	}
}

func TestGetCompleter(t *testing.T) {
	t.Parallel()

	completer := getCompleter(createMockCLIContext())
	assert.NotNil(t, completer, "Completer should not be nil")

	// Test that completer can be used without panicking
	assert.NotPanics(t, func() {
		completer.Do([]rune("help"), 0) // HACK:  zero was a wild guess
	})
}

func TestPrinter(t *testing.T) {
	t.Parallel()

	// Create a test writer to avoid nil pointer panics
	testWriter := &bytes.Buffer{}
	printer := &printer{out: testWriter}

	// Test that printer can handle different types without panicking
	testCases := []interface{}{
		nil,
		"hello",
		42,
		true,
		false,
		[]byte("bytes"),
	}

	for _, tc := range testCases {
		assert.NotPanics(t, func() {
			printer.Print(tc)
		}, "Printer should handle type %T without panicking", tc)
	}
}

func TestCommandStructure(t *testing.T) {
	t.Parallel()

	cmd := Command()

	// Test that command has expected properties
	assert.Equal(t, "shell", cmd.Name)
	assert.NotNil(t, cmd.Action)
	assert.NotNil(t, cmd.Before)
	assert.NotNil(t, cmd.After)
	assert.NotEmpty(t, cmd.Flags)

	// Test that expected flags are present
	flagNames := make([]string, len(cmd.Flags))
	for i, flag := range cmd.Flags {
		flagNames[i] = flag.Names()[0]
	}

	expectedFlags := []string{"ipfs", "command", "history-file", "prompt", "no-banner"}
	for _, expected := range expectedFlags {
		assert.Contains(t, flagNames, expected, "Command should have flag: %s", expected)
	}
}

func TestGlobalsIntegration(t *testing.T) {
	t.Parallel()

	baseGlobals := getBaseGlobals(createMockCLIContext())

	// Test that all expected functions are present and callable
	expectedFunctions := []string{"+", "*", ">", "<", "=", "/", "help", "println", "print"}

	for _, funcName := range expectedFunctions {
		funcVal, exists := baseGlobals[funcName]
		assert.True(t, exists, "Function %s should exist in base globals", funcName)
		assert.NotNil(t, funcVal, "Function %s should not be nil", funcName)
	}

	// Test that basic values are present
	expectedValues := []string{"nil", "true", "false", "version"}

	for _, valName := range expectedValues {
		val, exists := baseGlobals[valName]
		assert.True(t, exists, "Value %s should exist in base globals", valName)
		assert.NotNil(t, val, "Value %s should not be nil", valName)
	}
}

func TestArithmeticIntegration(t *testing.T) {
	t.Parallel()

	// Test that arithmetic works in command execution
	arithmeticTests := []struct {
		command     string
		expectError bool
	}{
		{"(+ 1 2)", false},
		{"(* 2 3)", false},
		{"(+ 1 2 3 4 5)", false},
		{"(* 2 3 4)", false},
		{"(> 5 3)", false},
		{"(< 3 5)", false},
		{"(= 5 5)", false},
		{"(* 10 0.5)", false},
	}

	for _, tt := range arithmeticTests {
		t.Run(tt.command, func(t *testing.T) {
			err := executeCommand(createMockCLIContext(), tt.command)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
