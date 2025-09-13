package shell

import (
	"context"
	"flag"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestGetBaseGlobals(t *testing.T) {
	t.Parallel()

	// Create a mock CLI context
	app := &cli.App{}
	app.Flags = []cli.Flag{}
	flagSet := &flag.FlagSet{}
	flagSet.Bool("with-console", true, "Enable console capability")
	flagSet.Bool("with-ipfs", false, "Enable IPFS capability") // Set to false for tests
	flagSet.Bool("with-exec", true, "Enable exec capability")
	flagSet.Bool("with-all", false, "Enable all capabilities")
	c := cli.NewContext(app, flagSet, nil)
	c.Context = context.Background()
	baseGlobals := getBaseGlobals(c)

	// Test that all expected base globals are present
	expectedGlobals := []string{
		"nil", "true", "false", "version",
		"=", "+", ">", "<", "*", "/",
		"help", "println", "print",
	}

	for _, expected := range expectedGlobals {
		assert.Contains(t, baseGlobals, expected, "Base globals should contain %s", expected)
	}

	// Test specific values
	assert.Equal(t, builtin.Nil{}, baseGlobals["nil"])
	assert.Equal(t, builtin.Bool(true), baseGlobals["true"])
	assert.Equal(t, builtin.Bool(false), baseGlobals["false"])
	assert.Equal(t, builtin.String("wetware-0.1.0"), baseGlobals["version"])
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

	// Test that arithmetic functions are present
	addFunc, exists := globals["+"]
	require.True(t, exists, "Addition function should be present")
	require.NotNil(t, addFunc, "Addition function should not be nil")

	mulFunc, exists := globals["*"]
	require.True(t, exists, "Multiplication function should be present")
	require.NotNil(t, mulFunc, "Multiplication function should not be nil")

	divFunc, exists := globals["/"]
	require.True(t, exists, "Division function should be present")
	require.NotNil(t, divFunc, "Division function should not be nil")
}

func TestComparisonFunctions(t *testing.T) {
	t.Parallel()

	// Test that comparison functions are present
	gtFunc, exists := globals[">"]
	require.True(t, exists, "Greater than function should be present")
	require.NotNil(t, gtFunc, "Greater than function should not be nil")

	ltFunc, exists := globals["<"]
	require.True(t, exists, "Less than function should be present")
	require.NotNil(t, ltFunc, "Less than function should not be nil")
}

func TestEqualityFunction(t *testing.T) {
	t.Parallel()

	// Test that equality function is present
	eqFunc, exists := globals["="]
	require.True(t, exists, "Equality function should be present")
	require.NotNil(t, eqFunc, "Equality function should not be nil")
}

func TestHelpFunction(t *testing.T) {
	t.Parallel()

	// Test that help function is present
	helpFunc, exists := globals["help"]
	require.True(t, exists, "Help function should be present")
	require.NotNil(t, helpFunc, "Help function should not be nil")
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
	app := &cli.App{}
	app.Flags = []cli.Flag{}
	flagSet := &flag.FlagSet{}
	flagSet.Bool("with-ipfs", false, "Enable IPFS capability") // Set to false for tests
	flagSet.Bool("with-exec", true, "Enable exec capability")
	flagSet.Bool("with-console", true, "Enable console capability")
	flagSet.Bool("with-all", false, "Enable all capabilities")
	c := cli.NewContext(app, flagSet, nil)
	c.Context = context.Background()
	baseGlobals1 := getBaseGlobals(c)
	baseGlobals2 := getBaseGlobals(c)

	// They should have the same content initially
	assert.Equal(t, baseGlobals1, baseGlobals2)

	// Modifying one shouldn't affect the other
	baseGlobals1["test"] = "modified"
	assert.NotContains(t, baseGlobals2, "test")
}
