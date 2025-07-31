package lang_test

import (
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
)

func TestDefaultEnvironment(t *testing.T) {
	t.Parallel()

	env := lang.DefaultEnvironment()

	// Test that basic Lisp functions are available
	require.Contains(t, env, "nil")
	require.Contains(t, env, "true")
	require.Contains(t, env, "false")

	// Test that list operations are available
	require.Contains(t, env, "cons")
	require.Contains(t, env, "conj")
	require.Contains(t, env, "list")
	require.Contains(t, env, "first")
	require.Contains(t, env, "rest")
	require.Contains(t, env, "count")

	// Test that comparison and arithmetic functions are available
	require.Contains(t, env, "=")
	require.Contains(t, env, "+")
	require.Contains(t, env, "-")
	require.Contains(t, env, "*")
	require.Contains(t, env, "/")
	require.Contains(t, env, ">")
	require.Contains(t, env, "<")
	require.Contains(t, env, ">=")
	require.Contains(t, env, "<=")

	// Test that utility functions are available
	require.Contains(t, env, "type")
	require.Contains(t, env, "str")
	// Note: println is not included in default environment - only added when console capability is available

	// Test that environment introspection functions are available
	require.Contains(t, env, "env")
	require.Contains(t, env, "help")
	require.Contains(t, env, "doc")
}

func TestBuiltinCons(t *testing.T) {
	t.Parallel()

	// Test cons with a list
	result, err := lang.BuiltinCons(builtin.String("a"), builtin.NewList(builtin.String("b"), builtin.String("c")))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test cons with a single value
	result, err = lang.BuiltinCons(builtin.String("a"), builtin.String("b"))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test cons with wrong number of arguments
	_, err = lang.BuiltinCons(builtin.String("a"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "cons requires exactly 2 arguments")

	_, err = lang.BuiltinCons(builtin.String("a"), builtin.String("b"), builtin.String("c"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "cons requires exactly 2 arguments")
}

func TestBuiltinConj(t *testing.T) {
	t.Parallel()

	// Test conj with a list
	list := builtin.NewList(builtin.String("a"), builtin.String("b"))
	result, err := lang.BuiltinConj(list, builtin.String("c"))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test conj with no arguments
	_, err = lang.BuiltinConj()
	require.Error(t, err)
	require.Contains(t, err.Error(), "conj requires at least 1 argument")
}

func TestBuiltinList(t *testing.T) {
	t.Parallel()

	// Test list with arguments
	result, err := lang.BuiltinList(builtin.String("a"), builtin.String("b"), builtin.String("c"))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test list with no arguments
	_, err = lang.BuiltinList()
	require.NoError(t, err)
	// builtin.NewList() with no args returns a nil LinkedList, which is valid for empty lists
	// We just check that no error was returned
}

func TestBuiltinFirst(t *testing.T) {
	t.Parallel()

	// Test first with a list
	list := builtin.NewList(builtin.String("a"), builtin.String("b"))
	result, err := lang.BuiltinFirst(list)
	require.NoError(t, err)
	require.Equal(t, builtin.String("a"), result)

	// Test first with wrong number of arguments
	_, err = lang.BuiltinFirst()
	require.Error(t, err)
	require.Contains(t, err.Error(), "first requires exactly 1 argument")

	_, err = lang.BuiltinFirst(list, builtin.String("extra"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "first requires exactly 1 argument")

	// Test first with non-sequence
	_, err = lang.BuiltinFirst(builtin.String("not-a-list"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "first argument must be a sequence")
}

func TestBuiltinRest(t *testing.T) {
	t.Parallel()

	// Test rest with a list
	list := builtin.NewList(builtin.String("a"), builtin.String("b"))
	result, err := lang.BuiltinRest(list)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test rest with wrong number of arguments
	_, err = lang.BuiltinRest()
	require.Error(t, err)
	require.Contains(t, err.Error(), "rest requires exactly 1 argument")

	// Test rest with non-sequence
	_, err = lang.BuiltinRest(builtin.String("not-a-list"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "first argument must be a sequence")
}

func TestBuiltinCount(t *testing.T) {
	t.Parallel()

	// Test count with a list
	list := builtin.NewList(builtin.String("a"), builtin.String("b"), builtin.String("c"))
	result, err := lang.BuiltinCount(list)
	require.NoError(t, err)
	require.Equal(t, builtin.Int64(3), result)

	// Test count with empty list
	emptyList := builtin.NewList()
	result, err = lang.BuiltinCount(emptyList)
	require.NoError(t, err)
	require.Equal(t, builtin.Int64(0), result)

	// Test count with non-sequence
	result, err = lang.BuiltinCount(builtin.String("not-a-list"))
	require.NoError(t, err)
	require.Equal(t, builtin.Int64(0), result)
}

func TestBuiltinEq(t *testing.T) {
	t.Parallel()

	// Test equality with same values
	result, err := lang.BuiltinEq(builtin.String("a"), builtin.String("a"))
	require.NoError(t, err)
	require.Equal(t, builtin.Bool(true), result)

	// Test equality with different values
	result, err = lang.BuiltinEq(builtin.String("a"), builtin.String("b"))
	require.NoError(t, err)
	require.Equal(t, builtin.Bool(false), result)

	// Test equality with wrong number of arguments
	_, err = lang.BuiltinEq(builtin.String("a"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "= requires exactly 2 arguments")
}

func TestBuiltinSum(t *testing.T) {
	t.Parallel()

	// Test sum with multiple arguments
	result, err := lang.BuiltinSum(builtin.Int64(1), builtin.Int64(2), builtin.Int64(3))
	require.NoError(t, err)
	require.Equal(t, builtin.Int64(6), result)

	// Test sum with no arguments
	result, err = lang.BuiltinSum()
	require.NoError(t, err)
	require.Equal(t, builtin.Int64(0), result)

	// Test sum with non-numeric arguments
	_, err = lang.BuiltinSum(builtin.String("not-a-number"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "+ requires numeric arguments")
}

func TestBuiltinType(t *testing.T) {
	t.Parallel()

	// Test type with string
	result, err := lang.BuiltinType(builtin.String("test"))
	require.NoError(t, err)
	require.Equal(t, builtin.String("builtin.String"), result)

	// Test type with number
	result, err = lang.BuiltinType(builtin.Int64(42))
	require.NoError(t, err)
	require.Equal(t, builtin.String("builtin.Int64"), result)

	// Test type with wrong number of arguments
	_, err = lang.BuiltinType()
	require.Error(t, err)
	require.Contains(t, err.Error(), "type requires exactly 1 argument")
}

func TestBuiltinStr(t *testing.T) {
	t.Parallel()

	// Test str with string
	result, err := lang.BuiltinStr(builtin.String("test"))
	require.NoError(t, err)
	require.Equal(t, builtin.String("test"), result)

	// Test str with number
	result, err = lang.BuiltinStr(builtin.Int64(42))
	require.NoError(t, err)
	require.Equal(t, builtin.String("42"), result)

	// Test str with wrong number of arguments
	_, err = lang.BuiltinStr()
	require.Error(t, err)
	require.Contains(t, err.Error(), "str requires exactly 1 argument")
}

func TestBuiltinEnv(t *testing.T) {
	t.Parallel()

	// Test env with no arguments
	result, err := lang.BuiltinEnv()
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test env with arguments
	_, err = lang.BuiltinEnv(builtin.String("extra"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "env takes no arguments")
}

func TestBuiltinHelp(t *testing.T) {
	t.Parallel()

	// Test help with no arguments
	result, err := lang.BuiltinHelp()
	require.NoError(t, err)
	require.NotNil(t, result)
	helpStr, ok := result.(builtin.String)
	require.True(t, ok)
	require.Contains(t, string(helpStr), "Wetware Shell")

	// Test help with function name
	result, err = lang.BuiltinHelp(builtin.String("cons"))
	require.NoError(t, err)
	require.NotNil(t, result)
	helpStr, ok = result.(builtin.String)
	require.True(t, ok)
	require.Contains(t, string(helpStr), "cons")

	// Test help with too many arguments
	_, err = lang.BuiltinHelp(builtin.String("cons"), builtin.String("extra"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "help takes 0 or 1 arguments")
}

func TestBuiltinDoc(t *testing.T) {
	t.Parallel()

	// Test doc with function name
	result, err := lang.BuiltinDoc(builtin.String("cons"))
	require.NoError(t, err)
	require.NotNil(t, result)
	docStr, ok := result.(builtin.String)
	require.True(t, ok)
	require.Contains(t, string(docStr), "Adds an element")

	// Test doc with wrong number of arguments
	_, err = lang.BuiltinDoc()
	require.Error(t, err)
	require.Contains(t, err.Error(), "doc requires exactly 1 argument")

	_, err = lang.BuiltinDoc(builtin.String("cons"), builtin.String("extra"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "doc requires exactly 1 argument")
}

// MockSession implements a mock session for testing NewGlobalsFromSession
type MockSession struct {
	hasConsole bool
	hasIpfs    bool
	hasExec    bool
	console    system.Console
	ipfs       system.IPFS
	exec       system.Executor
}

func (m *MockSession) HasConsole() bool        { return m.hasConsole }
func (m *MockSession) HasIpfs() bool           { return m.hasIpfs }
func (m *MockSession) HasExec() bool           { return m.hasExec }
func (m *MockSession) Console() system.Console { return m.console }
func (m *MockSession) Ipfs() system.IPFS       { return m.ipfs }
func (m *MockSession) Exec() system.Executor   { return m.exec }

func TestNewGlobalsFromSession(t *testing.T) {
	t.Parallel()

	// Test with no capabilities
	session := &MockSession{
		hasConsole: false,
		hasIpfs:    false,
		hasExec:    false,
	}

	env := lang.GlobalEnvConfig{
		Console:  session.Console(),
		IPFS:     session.Ipfs(),
		Executor: session.Exec(),
	}.New()
	require.NotNil(t, env)

	// Should have basic functions but no capability-specific ones
	require.Contains(t, env, "cons")
	require.Contains(t, env, "list")
	require.NotContains(t, env, "println") // No console capability
	require.NotContains(t, env, "ipfs")    // No IPFS capability
	require.NotContains(t, env, "go")      // No exec capability

	// Test with console capability only
	session.hasConsole = true
	// Note: We don't need to set actual console capability for this test
	// since we're only testing that the function is added to the environment

	env = lang.GlobalEnvConfig{
		Console:  session.Console(),
		IPFS:     session.Ipfs(),
		Executor: session.Exec(),
	}.New()
	require.NotNil(t, env)

	// Should have console function
	require.Contains(t, env, "println")
	require.NotContains(t, env, "ipfs")
	require.NotContains(t, env, "go")

	// Test with IPFS capability only
	session.hasConsole = false
	session.hasIpfs = true
	// Note: We don't need to set actual IPFS capability for this test
	// since we're only testing that the function is added to the environment

	env = lang.GlobalEnvConfig{
		Console:  session.Console(),
		IPFS:     session.Ipfs(),
		Executor: session.Exec(),
	}.New()
	require.NotNil(t, env)

	// Should have IPFS function
	require.Contains(t, env, "ipfs")
	require.NotContains(t, env, "println")
	require.NotContains(t, env, "go")

	// Test with exec capability only
	session.hasIpfs = false
	session.hasExec = true
	// Note: We don't need to set actual executor capability for this test
	// since we're only testing that the function is added to the environment

	env = lang.GlobalEnvConfig{
		Console:  session.Console(),
		IPFS:     session.Ipfs(),
		Executor: session.Exec(),
	}.New()
	require.NotNil(t, env)

	// Should have exec function
	require.Contains(t, env, "go")
	require.NotContains(t, env, "println")
	require.NotContains(t, env, "ipfs")

	// Test with all capabilities
	session.hasConsole = true
	session.hasIpfs = true
	session.hasExec = true

	env = lang.GlobalEnvConfig{
		Console:  session.Console(),
		IPFS:     session.Ipfs(),
		Executor: session.Exec(),
	}.New()
	require.NotNil(t, env)

	// Should have all functions
	require.Contains(t, env, "println")
	require.Contains(t, env, "ipfs")
	require.Contains(t, env, "go")
}
