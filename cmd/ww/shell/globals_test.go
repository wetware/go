package shell

import (
	"context"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBaseGlobals(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseGlobals := getBaseGlobals(ctx)

	// Test that all expected base globals are present
	expectedGlobals := []string{
		"nil", "true", "false", "version",
		"=", "+", ">", "<", "*", "/",
		"help", "println", "print",
		"ipfs",
	}

	for _, expected := range expectedGlobals {
		assert.Contains(t, baseGlobals, expected, "Base globals should contain %s", expected)
	}

	// Test specific values
	assert.Equal(t, builtin.Nil{}, baseGlobals["nil"])
	assert.Equal(t, builtin.Bool(true), baseGlobals["true"])
	assert.Equal(t, builtin.Bool(false), baseGlobals["false"])
	assert.Equal(t, builtin.String("wetware-0.1.0"), baseGlobals["version"])

	// Test that ipfs is present (it should be a function)
	assert.NotNil(t, baseGlobals["ipfs"])
}

func TestGlobalsFromGlobalsGo(t *testing.T) {
	t.Parallel()

	// Test that the globals map from globals.go has expected content
	expectedGlobals := []string{
		"nil", "true", "false", "version",
		"=", "+", ">", "<", "*", "/",
		"help", "println", "print",
	}

	for _, expected := range expectedGlobals {
		assert.Contains(t, globals, expected, "globals map should contain %s", expected)
	}

	// Test specific values
	assert.Equal(t, builtin.Nil{}, globals["nil"])
	assert.Equal(t, builtin.Bool(true), globals["true"])
	assert.Equal(t, builtin.Bool(false), globals["false"])
	assert.Equal(t, builtin.String("wetware-0.1.0"), globals["version"])
}

func TestArithmeticFunctions(t *testing.T) {
	t.Parallel()

	// Test addition function
	addFunc, ok := globals["+"].(func(...int) int)
	require.True(t, ok, "Addition function should be present and callable")

	result := addFunc(1, 2, 3, 4)
	assert.Equal(t, 10, result)

	// Test multiplication function
	mulFunc, ok := globals["*"].(func(...int) int)
	require.True(t, ok, "Multiplication function should be present and callable")

	result = mulFunc(2, 3, 4)
	assert.Equal(t, 24, result)

	// Test division function
	divFunc, ok := globals["/"].(func(builtin.Int64, builtin.Int64) float64)
	require.True(t, ok, "Division function should be present and callable")

	resultFloat := divFunc(10, 2)
	assert.Equal(t, 5.0, resultFloat)
}

func TestComparisonFunctions(t *testing.T) {
	t.Parallel()

	// Test greater than function
	gtFunc, ok := globals[">"].(func(builtin.Int64, builtin.Int64) bool)
	require.True(t, ok, "Greater than function should be present and callable")

	assert.True(t, gtFunc(5, 3))
	assert.False(t, gtFunc(3, 5))
	assert.False(t, gtFunc(5, 5))

	// Test less than function
	ltFunc, ok := globals["<"].(func(builtin.Int64, builtin.Int64) bool)
	require.True(t, ok, "Less than function should be present and callable")

	assert.True(t, ltFunc(3, 5))
	assert.False(t, ltFunc(5, 3))
	assert.False(t, ltFunc(5, 5))
}

func TestEqualityFunction(t *testing.T) {
	t.Parallel()

	// Test equality function
	eqFunc, ok := globals["="].(func(interface{}, interface{}) bool)
	require.True(t, ok, "Equality function should be present and callable")

	assert.True(t, eqFunc(5, 5))
	assert.True(t, eqFunc("hello", "hello"))
	assert.False(t, eqFunc(5, 6))
	assert.False(t, eqFunc("hello", "world"))
}

func TestHelpFunction(t *testing.T) {
	t.Parallel()

	// Test help function
	helpFunc, ok := globals["help"].(func() string)
	require.True(t, ok, "Help function should be present and callable")

	helpText := helpFunc()
	assert.Contains(t, helpText, "Wetware Shell")
	assert.Contains(t, helpText, "help")
	assert.Contains(t, helpText, "version")
	assert.Contains(t, helpText, "println")
	assert.Contains(t, helpText, "IPFS Path Syntax")
}

func TestPrintFunctions(t *testing.T) {
	t.Parallel()

	// Test that println function exists
	printlnFunc, exists := globals["println"]
	assert.True(t, exists, "Println function should exist")
	assert.NotNil(t, printlnFunc, "Println function should not be nil")

	// Test that print function exists
	printFunc, exists := globals["print"]
	assert.True(t, exists, "Print function should exist")
	assert.NotNil(t, printFunc, "Print function should not be nil")
}

func TestGlobalsConsistency(t *testing.T) {
	t.Parallel()

	// Test that getBaseGlobals returns a copy, not the original
	ctx := context.Background()
	baseGlobals1 := getBaseGlobals(ctx)
	baseGlobals2 := getBaseGlobals(ctx)

	// They should be different maps (different memory addresses)
	assert.NotSame(t, baseGlobals1, baseGlobals2)

	// But they should have the same content
	assert.Equal(t, baseGlobals1, baseGlobals2)

	// Modifying one shouldn't affect the other
	baseGlobals1["test"] = "value"
	assert.NotEqual(t, baseGlobals1, baseGlobals2)
	assert.NotContains(t, baseGlobals2, "test")
}
