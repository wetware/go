package shell

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

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
			err := executeCommand(ctx, tt.command)

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

	ctx := context.Background()

	// These tests will fail if IPFS is not available, but they test the structure
	tests := []struct {
		name      string
		command   string
		wantError bool
	}{
		{
			name:      "ipfs function exists",
			command:   "ipfs",
			wantError: false,
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
			err := executeCommand(ctx, tt.command)

			if tt.wantError {
				assert.Error(t, err, "Expected error for command: %s", tt.command)
			} else {
				assert.NoError(t, err, "Expected no error for command: %s", tt.command)
			}
		})
	}
}

func TestGetReadlineConfig(t *testing.T) {
	t.Parallel()

	// This is a basic test to ensure the function doesn't panic
	// and returns a valid config
	config := getReadlineConfig(nil)

	// Test that config has expected fields
	assert.NotEmpty(t, config.Prompt)
	assert.NotEmpty(t, config.HistoryFile)
	assert.NotNil(t, config.AutoComplete)
	assert.Equal(t, "^C", config.InterruptPrompt)
	assert.Equal(t, "exit", config.EOFPrompt)
}

func TestGetCompleter(t *testing.T) {
	t.Parallel()

	completer := getCompleter()
	assert.NotNil(t, completer, "Completer should not be nil")

	// Test that completer can be used without panicking
	assert.NotPanics(t, func() {
		completer.Do([]rune("help"), 0) // HACK:  zero was a wild guess
	})
}

func TestPrinter(t *testing.T) {
	t.Parallel()

	printer := printer{out: nil} // We can't easily test output without capturing it

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

	ctx := context.Background()
	baseGlobals := getBaseGlobals(ctx)

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

	ctx := context.Background()

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
		{"(/ 10 2)", false},
	}

	for _, tt := range arithmeticTests {
		t.Run(tt.command, func(t *testing.T) {
			err := executeCommand(ctx, tt.command)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
