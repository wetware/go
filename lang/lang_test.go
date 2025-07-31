package lang_test

import (
	"context"
	"errors"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
)

// TestIPFSCat tests the standalone IPFSCat function
func TestIPFSCat(t *testing.T) {
	t.Parallel()
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}
	mock := system.IPFS_ServerToClient(mockServer)

	// Test with UnixPath
	unixPath, err := lang.NewUnixPath("/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa")
	require.NoError(t, err, "Failed to create UnixPath")

	result, err := lang.IPFSCat(mock, unixPath)
	require.NoError(t, err, "Failed to invoke cat with UnixPath")

	buffer, ok := result.(*lang.Buffer)
	require.True(t, ok, "Expected *lang.Buffer result, got %T", result)

	require.Equal(t, "0x746573742064617461", buffer.AsHex(), "Buffer hex representation mismatch")

	t.Logf("Successfully tested IPFSCat with UnixPath argument")
}

// TestBuffer tests the Buffer type directly
func TestBuffer(t *testing.T) {
	t.Parallel()
	// Test empty buffer
	emptyBuffer := &lang.Buffer{}
	require.Equal(t, "", emptyBuffer.String(), "Empty buffer should return ''")
	require.Equal(t, "0x", emptyBuffer.AsHex(), "Empty buffer should return '0x'")

	// Test buffer with data
	testData := []byte("test data")
	buffer := &lang.Buffer{Mem: testData}
	require.Equal(t, "test data", buffer.String(), "Buffer string representation mismatch")
	require.Equal(t, "0x746573742064617461", buffer.AsHex(), "Buffer hex representation mismatch")
}

// TestIPFSAdd tests the IPFSAdd function
func TestIPFSAdd(t *testing.T) {
	t.Parallel()
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}
	mock := system.IPFS_ServerToClient(mockServer)

	// Create a Buffer with test data
	testData := []byte("test data for add")
	buffer := &lang.Buffer{Mem: testData}

	// Test with Buffer
	result, err := lang.IPFSAdd(mock, buffer)
	require.NoError(t, err, "Failed to invoke add with Buffer")

	cid, ok := result.(string)
	require.True(t, ok, "Expected string result, got %T", result)

	require.Equal(t, "QmTest123", cid, "CID mismatch")

	t.Logf("Successfully tested IPFSAdd with Buffer argument")
}

// TestGoSpecialForm tests the Go special form argument validation
func TestGoSpecialForm(t *testing.T) {
	t.Parallel()
	// Test argument validation
	goForm := lang.Go{}

	// Test with insufficient arguments
	_, err := goForm.Invoke()
	if err == nil {
		t.Error("Expected error for insufficient arguments")
	}

	// Test with wrong first argument type
	_, err = goForm.Invoke("not-a-path", builtin.NewList())
	if err == nil {
		t.Error("Expected error for wrong first argument type")
	}
}

// TestGoSpecialFormWithMockExecutor tests that the go special form correctly calls the executor
func TestGoSpecialFormWithMockExecutor(t *testing.T) {
	t.Parallel()
	mockExecutor := &MockExecutor{
		spawnCalled:   false,
		spawnPath:     "",
		spawnArgs:     nil,
		spawnDir:      "",
		spawnEnv:      nil,
		shouldSucceed: true,
	}

	// Convert mock to client
	executorClient := system.Executor_ServerToClient(mockExecutor)
	defer executorClient.Release()

	// Create the Go special form with the mock executor
	goForm := lang.Go{Executor: executorClient}

	// Create test arguments
	execPath, err := lang.NewUnixPath("/ipfs/QmWKKmjmTmbaFuU4Bu92KXob3jaqKJ9vZXRch6Ks8GJESZ/cmd/shell")
	require.NoError(t, err, "Failed to create UnixPath")

	body := builtin.String("(console.Println \"Hello, World!\")")

	// Test basic invocation
	_, err = goForm.Invoke(execPath, body)
	require.NoError(t, err, "Go special form should not return error")
	require.True(t, mockExecutor.spawnCalled, "Executor.Spawn should have been called")
	require.Equal(t, "/ipfs/QmWKKmjmTmbaFuU4Bu92KXob3jaqKJ9vZXRch6Ks8GJESZ/cmd/shell", mockExecutor.spawnPath, "Path should match")
	require.Equal(t, "(console.Println \"Hello, World!\")", mockExecutor.spawnArgs[0], "First argument should be the body")
	require.Equal(t, "", mockExecutor.spawnDir, "Working directory should be empty")

	// Test with keyword arguments
	mockExecutor.spawnCalled = false
	mockExecutor.spawnArgs = nil

	// Create keyword arguments
	_, err = goForm.Invoke(execPath, body,
		builtin.String("console"), builtin.String("test-console"),
		builtin.String("data"), builtin.String("test-data"))
	require.NoError(t, err, "Go special form with kwargs should not return error")
	require.True(t, mockExecutor.spawnCalled, "Executor.Spawn should have been called")
	require.Equal(t, "(console.Println \"Hello, World!\")", mockExecutor.spawnArgs[0], "First argument should be the body")
	require.Equal(t, "--data=\"test-data\"", mockExecutor.spawnArgs[1], "Second argument should be data kwarg")
	require.Equal(t, 2, len(mockExecutor.spawnArgs), "Should have exactly 2 arguments (body + data kwarg)")

	// Test error case
	mockExecutor.shouldSucceed = false
	mockExecutor.spawnCalled = false
	mockExecutor.spawnArgs = nil

	_, err = goForm.Invoke(execPath, body)
	require.Error(t, err, "Go special form should return error when spawn fails")
	require.True(t, mockExecutor.spawnCalled, "Executor.Spawn should have been called even on error")
}

// MockIPFSServer implements system.IPFS_Server for testing
type MockIPFSServer struct {
	testValue int
}

// Add implements system.IPFS_Server.Add
func (m *MockIPFSServer) Add(ctx context.Context, call system.IPFS_add) error {
	// Mock implementation - just return a dummy CID
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	return results.SetCid("QmTest123")
}

// Cat implements system.IPFS_Server.Cat
func (m *MockIPFSServer) Cat(ctx context.Context, call system.IPFS_cat) error {
	// Mock implementation - return some test data
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	testData := []byte("test data")
	return results.SetBody(testData)
}

// Ls implements system.IPFS_Server.Ls
func (m *MockIPFSServer) Ls(ctx context.Context, call system.IPFS_ls) error {
	// Mock implementation - return empty list
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	entries, err := results.NewEntries(0)
	if err != nil {
		return err
	}
	return results.SetEntries(entries)
}

// Stat implements system.IPFS_Server.Stat
func (m *MockIPFSServer) Stat(ctx context.Context, call system.IPFS_stat) error {
	// Mock implementation
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	info, err := results.NewInfo()
	if err != nil {
		return err
	}
	info.SetCid("QmTest123")
	info.SetSize(100)
	info.SetCumulativeSize(100)
	_, err = info.NodeType().NewFile()
	if err != nil {
		return err
	}
	return results.SetInfo(info)
}

// Pin implements system.IPFS_Server.Pin
func (m *MockIPFSServer) Pin(ctx context.Context, call system.IPFS_pin) error {
	// Mock implementation - always succeed
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	results.SetSuccess(true)
	return nil
}

// Unpin implements system.IPFS_Server.Unpin
func (m *MockIPFSServer) Unpin(ctx context.Context, call system.IPFS_unpin) error {
	// Mock implementation - always succeed
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	results.SetSuccess(true)
	return nil
}

// Pins implements system.IPFS_Server.Pins
func (m *MockIPFSServer) Pins(ctx context.Context, call system.IPFS_pins) error {
	// Mock implementation - return some test CIDs
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	cids, err := results.NewCids(2)
	if err != nil {
		return err
	}
	cids.Set(0, "QmTest1")
	cids.Set(1, "QmTest2")
	return results.SetCids(cids)
}

// Id implements system.IPFS_Server.Id
func (m *MockIPFSServer) Id(ctx context.Context, call system.IPFS_id) error {
	// Mock implementation
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	info, err := results.NewPeerInfo()
	if err != nil {
		return err
	}
	info.SetId("QmTestPeer")
	addresses, err := info.NewAddresses(1)
	if err != nil {
		return err
	}
	addresses.Set(0, "/ip4/127.0.0.1/tcp/4001")
	return results.SetPeerInfo(info)
}

// Connect implements system.IPFS_Server.Connect
func (m *MockIPFSServer) Connect(ctx context.Context, call system.IPFS_connect) error {
	// Mock implementation - always succeed
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	results.SetSuccess(true)
	return nil
}

// Peers implements system.IPFS_Server.Peers
func (m *MockIPFSServer) Peers(ctx context.Context, call system.IPFS_peers) error {
	// Mock implementation - return empty list
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	peerList, err := results.NewPeerList(0)
	if err != nil {
		return err
	}
	return results.SetPeerList(peerList)
}

// ResolveNode implements system.IPFS_Server.ResolveNode
func (m *MockIPFSServer) ResolveNode(ctx context.Context, call system.IPFS_resolveNode) error {
	// Mock implementation - return test CID
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	return results.SetCid("QmTestResolved")
}

// MockExecutor implements system.Executor_Server for testing
type MockExecutor struct {
	spawnCalled   bool
	spawnPath     string
	spawnArgs     []string
	spawnDir      string
	spawnEnv      []string
	shouldSucceed bool
}

// Spawn implements system.Executor_Server.Spawn
func (m *MockExecutor) Spawn(ctx context.Context, call system.Executor_spawn) error {
	m.spawnCalled = true

	// Get the parameters
	params := call.Args()

	// Extract command
	command, err := params.Command()
	if err != nil {
		return err
	}

	// Extract path
	path, err := command.Path()
	if err != nil {
		return err
	}
	m.spawnPath = path

	// Extract arguments
	args, err := command.Args()
	if err != nil {
		return err
	}
	m.spawnArgs = make([]string, args.Len())
	for i := 0; i < args.Len(); i++ {
		arg, err := args.At(i)
		if err != nil {
			return err
		}
		m.spawnArgs[i] = arg
	}

	// Extract working directory
	dir, err := command.Dir()
	if err != nil {
		return err
	}
	m.spawnDir = dir

	// Extract environment variables
	env, err := command.Env()
	if err != nil {
		return err
	}
	m.spawnEnv = make([]string, env.Len())
	for i := 0; i < env.Len(); i++ {
		envVar, err := env.At(i)
		if err != nil {
			return err
		}
		m.spawnEnv[i] = envVar
	}

	// Allocate results
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	if m.shouldSucceed {
		// Create a successful result
		cell := system.Cell_ServerToClient(&MockCell{})
		optionalCell, err := results.NewCell()
		if err != nil {
			return err
		}
		return optionalCell.SetCell(cell)
	} else {
		// Return an error
		return errors.New("mock executor spawn failed")
	}
}

// MockCell implements system.Cell_Server for testing
type MockCell struct{}

// Wait implements system.Cell_Server.Wait
func (m *MockCell) Wait(ctx context.Context, call system.Cell_wait) error {
	// Mock implementation - return success
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	// Create a successful result
	result, err := results.NewResult()
	if err != nil {
		return err
	}
	result.SetOk()

	return results.SetResult(result)
}

// TestConsolePrintln tests the ConsolePrintln command
func TestConsolePrintln(t *testing.T) {
	t.Parallel()
	// Create a mock console server
	mockConsole := &MockConsoleServer{
		output:       "",
		bytesWritten: 0,
	}
	consoleClient := system.Console_ServerToClient(mockConsole)
	defer consoleClient.Release()

	// Create the ConsolePrintln function
	printlnFunc := lang.ConsolePrintln{Console: consoleClient}

	// Test with string argument
	result, err := printlnFunc.Invoke("Hello, Wetware!")
	require.NoError(t, err, "Failed to invoke println with string")

	bytesWritten, ok := result.(builtin.Int64)
	require.True(t, ok, "Expected builtin.Int64 result, got %T", result)
	require.Equal(t, builtin.Int64(16), bytesWritten, "Expected 16 bytes written (including newline)")
	require.Equal(t, "Hello, Wetware!", mockConsole.output, "Console should have received the correct output")

	// Test with builtin.String argument
	mockConsole.output = ""
	mockConsole.bytesWritten = 0

	result, err = printlnFunc.Invoke(builtin.String("Test String"))
	require.NoError(t, err, "Failed to invoke println with builtin.String")

	bytesWritten, ok = result.(builtin.Int64)
	require.True(t, ok, "Expected builtin.Int64 result, got %T", result)
	require.Equal(t, builtin.Int64(12), bytesWritten, "Expected 12 bytes written (including newline)")
	require.Equal(t, "Test String", mockConsole.output, "Console should have received the correct output")

	// Test with Buffer argument
	mockConsole.output = ""
	mockConsole.bytesWritten = 0

	buffer := &lang.Buffer{Mem: []byte("Buffer content")}
	result, err = printlnFunc.Invoke(buffer)
	require.NoError(t, err, "Failed to invoke println with Buffer")

	bytesWritten, ok = result.(builtin.Int64)
	require.True(t, ok, "Expected builtin.Int64 result, got %T", result)
	require.Equal(t, builtin.Int64(15), bytesWritten, "Expected 15 bytes written (including newline)")
	require.Equal(t, "Buffer content", mockConsole.output, "Console should have received the correct output")

	// Test with no arguments (identity law)
	result, err = printlnFunc.Invoke()
	require.NoError(t, err, "Failed to invoke println with no arguments")

	// Should return the function itself
	returnedFunc, ok := result.(lang.ConsolePrintln)
	require.True(t, ok, "Expected ConsolePrintln result, got %T", result)
	require.Equal(t, consoleClient, returnedFunc.Console, "Returned function should have the same console")

	// Test with wrong number of arguments
	_, err = printlnFunc.Invoke("arg1", "arg2")
	require.Error(t, err, "Expected error for wrong number of arguments")
	require.Contains(t, err.Error(), "println requires exactly 1 argument, got 2", "Error message should be descriptive")

	// Test with unsupported argument type
	_, err = printlnFunc.Invoke(42)
	require.NoError(t, err, "Should handle unsupported types gracefully")
	require.Equal(t, "42", mockConsole.output, "Console should have received string representation of number")
}

// MockConsoleServer implements system.Console_Server for testing
type MockConsoleServer struct {
	output       string
	bytesWritten uint32
}

// Println implements system.Console_Server.Println
func (m *MockConsoleServer) Println(ctx context.Context, call system.Console_println) error {
	// Get the output data from the call
	output, err := call.Args().Output()
	if err != nil {
		return err
	}

	// Store the output for testing
	m.output = output
	m.bytesWritten = uint32(len(output) + 1) // +1 for newline

	// Set the result (number of bytes written)
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	results.SetN(m.bytesWritten)

	return nil
}

// TestDotNotationAnalyzer tests the DotNotationAnalyzer functionality
func TestDotNotationAnalyzer(t *testing.T) {
	t.Parallel()

	// Create a mock base analyzer for testing
	mockBase := &MockAnalyzer{
		analyzedForms: make([]any, 0),
	}

	// Create the dot notation analyzer
	analyzer := lang.NewDotNotationAnalyzer(mockBase)

	// Test cases
	testCases := []struct {
		name        string
		input       any
		expected    any
		expectError bool
		description string
	}{
		{
			name:        "Basic dot notation transformation",
			input:       createTestList(builtin.Symbol("ipfs.stat"), "/ipfs/QmTest"),
			expected:    createTestList(builtin.Symbol("ipfs"), builtin.String("stat"), "/ipfs/QmTest"),
			expectError: false,
			description: "Should transform (ipfs.stat path) to (ipfs \"stat\" path)",
		},
		{
			name:        "Dot notation with multiple arguments",
			input:       createTestList(builtin.Symbol("ipfs.cat"), "/ipfs/QmTest", "arg2", "arg3"),
			expected:    createTestList(builtin.Symbol("ipfs"), builtin.String("cat"), "/ipfs/QmTest", "arg2", "arg3"),
			expectError: false,
			description: "Should preserve all arguments after transformation",
		},
		{
			name:        "Dot notation with no arguments",
			input:       createTestList(builtin.Symbol("ipfs.id")),
			expected:    createTestList(builtin.Symbol("ipfs"), builtin.String("id")),
			expectError: false,
			description: "Should handle method calls with no arguments",
		},
		{
			name:        "Complex object dot notation",
			input:       createTestList(builtin.Symbol("myObject.myMethod"), "param1", 42, true),
			expected:    createTestList(builtin.Symbol("myObject"), builtin.String("myMethod"), "param1", 42, true),
			expectError: false,
			description: "Should work with any object.method pattern",
		},
		{
			name:        "Non-dot notation symbol",
			input:       createTestList(builtin.Symbol("regularFunction"), "arg1", "arg2"),
			expected:    createTestList(builtin.Symbol("regularFunction"), "arg1", "arg2"),
			expectError: false,
			description: "Should pass through non-dot notation unchanged",
		},
		{
			name:        "Empty list",
			input:       createTestList(),
			expected:    createTestList(),
			expectError: false,
			description: "Should handle empty lists",
		},
		{
			name:        "Non-symbol first element",
			input:       createTestList(42, "not-a-symbol"),
			expected:    createTestList(42, "not-a-symbol"),
			expectError: false,
			description: "Should pass through when first element is not a symbol",
		},
		{
			name:        "Symbol without dot",
			input:       createTestList(builtin.Symbol("noDotSymbol"), "arg1"),
			expected:    createTestList(builtin.Symbol("noDotSymbol"), "arg1"),
			expectError: false,
			description: "Should pass through symbols without dots",
		},
		{
			name:        "Multiple dots in symbol",
			input:       createTestList(builtin.Symbol("obj.method.submethod"), "arg1"),
			expected:    createTestList(builtin.Symbol("obj"), builtin.String("method.submethod"), "arg1"),
			expectError: false,
			description: "Should split only on first dot",
		},
		{
			name:        "Non-list input",
			input:       "not-a-list",
			expected:    "not-a-list",
			expectError: false,
			description: "Should pass through non-list inputs unchanged",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the mock analyzer
			mockBase.analyzedForms = make([]any, 0)

			// Call the analyzer
			expr, err := analyzer.Analyze(nil, tc.input)

			// Check error expectations
			if tc.expectError {
				require.Error(t, err, tc.description)
				return
			}
			require.NoError(t, err, tc.description)

			// Verify that the base analyzer was called with the expected form
			require.Len(t, mockBase.analyzedForms, 1, "Base analyzer should be called exactly once")
			require.Equal(t, tc.expected, mockBase.analyzedForms[0], tc.description)

			// Verify the returned expression is not nil
			require.NotNil(t, expr, "Analyzer should return a non-nil expression")
		})
	}
}

// TestDotNotationAnalyzerErrorHandling tests error handling in DotNotationAnalyzer
func TestDotNotationAnalyzerErrorHandling(t *testing.T) {
	t.Parallel()

	// Create a mock base analyzer that returns errors
	mockBase := &MockAnalyzer{
		shouldError: true,
		errorMsg:    "base analyzer error",
	}

	analyzer := lang.NewDotNotationAnalyzer(mockBase)

	// Test that errors from base analyzer are propagated
	_, err := analyzer.Analyze(nil, createTestList(builtin.Symbol("ipfs.stat"), "arg1"))
	require.Error(t, err, "Should propagate errors from base analyzer")
	require.Contains(t, err.Error(), "base analyzer error", "Should preserve original error message")
}

// TestDotNotationAnalyzerNilBase tests behavior with nil base analyzer
func TestDotNotationAnalyzerNilBase(t *testing.T) {
	t.Parallel()

	// Create analyzer with nil base (should use default builtin analyzer)
	analyzer := lang.NewDotNotationAnalyzer(nil)

	// Test that it still works with dot notation
	expr, err := analyzer.Analyze(nil, createTestList(builtin.Symbol("ipfs.stat"), "/ipfs/QmTest"))
	require.NoError(t, err, "Should work with nil base analyzer")
	require.NotNil(t, expr, "Should return non-nil expression")
}

// TestDotNotationAnalyzerEdgeCases tests edge cases and boundary conditions
func TestDotNotationAnalyzerEdgeCases(t *testing.T) {
	t.Parallel()

	mockBase := &MockAnalyzer{analyzedForms: make([]any, 0)}
	analyzer := lang.NewDotNotationAnalyzer(mockBase)

	edgeCases := []struct {
		name        string
		input       any
		description string
	}{
		{
			name:        "Symbol starting with dot",
			input:       createTestList(builtin.Symbol(".method"), "arg1"),
			description: "Should handle symbols starting with dot",
		},
		{
			name:        "Symbol ending with dot",
			input:       createTestList(builtin.Symbol("object."), "arg1"),
			description: "Should handle symbols ending with dot",
		},
		{
			name:        "Symbol with only dot",
			input:       createTestList(builtin.Symbol("."), "arg1"),
			description: "Should handle symbol with only dot",
		},
		{
			name:        "Multiple consecutive dots",
			input:       createTestList(builtin.Symbol("obj..method"), "arg1"),
			description: "Should handle multiple consecutive dots",
		},
		{
			name:        "Nil input",
			input:       nil,
			description: "Should handle nil input gracefully",
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			mockBase.analyzedForms = make([]any, 0)

			// Should not panic and should delegate to base analyzer
			expr, err := analyzer.Analyze(nil, tc.input)

			// These edge cases should be handled by the base analyzer
			// We just verify they don't cause panics
			if err == nil {
				require.NotNil(t, expr, tc.description)
			}
		})
	}
}

// TestDotNotationAnalyzerIntegration tests integration with real builtin analyzer
func TestDotNotationAnalyzerIntegration(t *testing.T) {
	t.Parallel()

	// Create analyzer with real builtin analyzer
	analyzer := lang.NewDotNotationAnalyzer(&builtin.Analyzer{})

	// Test that it can analyze dot notation expressions
	expr, err := analyzer.Analyze(nil, createTestList(builtin.Symbol("ipfs.stat"), "/ipfs/QmTest"))
	require.NoError(t, err, "Should work with real builtin analyzer")
	require.NotNil(t, expr, "Should return non-nil expression")

	// Test that the expression can be evaluated (basic smoke test)
	// Note: We can't actually evaluate it without a proper environment,
	// but we can verify it's a valid expression
	require.NotNil(t, expr, "Expression should be valid")
}

// TestNewIPLDLinkedList tests the NewIPLDLinkedList function
func TestNewIPLDLinkedList(t *testing.T) {
	// Create a mock IPFS capability
	mockIPFS := system.IPFS{}

	// Test with some values
	values := []core.Any{
		builtin.String("first"),
		builtin.String("second"),
		builtin.String("third"),
	}

	// Create the IPLD linked list
	list, err := lang.NewIPLDLinkedList(mockIPFS, values...)
	require.NoError(t, err)
	require.NotNil(t, list)

	// Verify the list properties
	assert.Equal(t, 3, list.GetIPLDElementCount())
	assert.NotEmpty(t, list.GetIPLDHeadCID())

	// Verify the builtin linked list compatibility
	count, err := list.Count()
	require.NoError(t, err)
	assert.Equal(t, 3, count)
	first, err := list.First()
	require.NoError(t, err)
	assert.Equal(t, builtin.String("first"), first)
}

// TestNewIPLDLinkedListEmpty tests creating an empty IPLD linked list
func TestNewIPLDLinkedListEmpty(t *testing.T) {
	t.Parallel()
	mockIPFS := system.IPFS{}

	// Test with no values
	list, err := lang.NewIPLDLinkedList(mockIPFS)
	require.NoError(t, err)
	require.NotNil(t, list)

	// Verify empty list properties
	assert.Equal(t, 0, list.GetIPLDElementCount())
	assert.Empty(t, list.GetIPLDHeadCID())

	// Test First() on empty list
	_, err = list.First()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty list")

	// Test Count() on empty list
	count, err := list.Count()
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Test SExpr() on empty list
	sexpr, err := list.SExpr()
	require.NoError(t, err)
	assert.Equal(t, "()", sexpr)
}

// TestIPLDLinkedListMethods tests the various methods of IPLDLinkedList
func TestIPLDLinkedListMethods(t *testing.T) {
	t.Parallel()
	mockIPFS := system.IPFS{}

	// Test with single value
	list, err := lang.NewIPLDLinkedList(mockIPFS, builtin.String("single"))
	require.NoError(t, err)

	// Test GetIPLDHeadCID
	headCID := list.GetIPLDHeadCID()
	assert.NotEmpty(t, headCID)

	// Test GetIPLDElementCount
	count := list.GetIPLDElementCount()
	assert.Equal(t, 1, count)

	// Test First
	first, err := list.First()
	require.NoError(t, err)
	assert.Equal(t, builtin.String("single"), first)

	// Test Next (should return nil for single element)
	next, err := list.Next()
	require.NoError(t, err)
	assert.Nil(t, next)

	// Test Count
	countResult, err := list.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, countResult)

	// Test SExpr
	sexpr, err := list.SExpr()
	require.NoError(t, err)
	assert.Equal(t, "(\"single\")", sexpr)
}

// TestIPLDLinkedListConj tests the Conj method
func TestIPLDLinkedListConj(t *testing.T) {
	t.Parallel()
	mockIPFS := system.IPFS{}

	// Create initial list
	list, err := lang.NewIPLDLinkedList(mockIPFS, builtin.String("original"))
	require.NoError(t, err)

	// Test Conj with additional items
	newList, err := list.Conj(builtin.String("added1"), builtin.String("added2"))
	require.NoError(t, err)
	require.NotNil(t, newList)

	// Verify the new list has the correct structure
	// The Conj method prepends the current value to the new items
	// So we expect: ["original", "added1", "added2"]
	count, err := newList.Count()
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	first, err := newList.First()
	require.NoError(t, err)
	assert.Equal(t, builtin.String("original"), first)
}

// TestIPLDLinkedListConjEmpty tests Conj on empty list
func TestIPLDLinkedListConjEmpty(t *testing.T) {
	t.Parallel()
	mockIPFS := system.IPFS{}

	// Create empty list
	list, err := lang.NewIPLDLinkedList(mockIPFS)
	require.NoError(t, err)

	// Test Conj on empty list
	newList, err := list.Conj(builtin.String("item1"), builtin.String("item2"))
	require.NoError(t, err)
	require.NotNil(t, newList)

	// Since the original list was empty, Conj should create a list with just the new items
	count, err := newList.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count) // ["item1", "item2"]

	// The first element should be the first new item
	first, err := newList.First()
	require.NoError(t, err)
	assert.Equal(t, builtin.String("item1"), first)
}

// TestBuiltinArithmetic tests the arithmetic builtin functions
func TestBuiltinArithmetic(t *testing.T) {
	t.Parallel()
	// Test BuiltinSub (subtraction)
	subResult, err := lang.BuiltinSub(builtin.Int64(10), builtin.Int64(3))
	require.NoError(t, err)
	assert.Equal(t, builtin.Int64(7), subResult)

	// Test BuiltinMul (multiplication)
	mulResult, err := lang.BuiltinMul(builtin.Int64(4), builtin.Int64(5))
	require.NoError(t, err)
	assert.Equal(t, builtin.Int64(20), mulResult)

	// Test BuiltinDiv (division)
	divResult, err := lang.BuiltinDiv(builtin.Int64(15), builtin.Int64(3))
	require.NoError(t, err)
	assert.Equal(t, builtin.Int64(5), divResult)

	// Test division by zero
	_, err = lang.BuiltinDiv(builtin.Int64(10), builtin.Int64(0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero")
}

// TestBuiltinComparison tests the comparison builtin functions
func TestBuiltinComparison(t *testing.T) {
	t.Parallel()
	// Test BuiltinGt (greater than)
	gtResult, err := lang.BuiltinGt(builtin.Int64(10), builtin.Int64(5))
	require.NoError(t, err)
	assert.Equal(t, builtin.Bool(true), gtResult)

	gtResult, err = lang.BuiltinGt(builtin.Int64(5), builtin.Int64(10))
	require.NoError(t, err)
	assert.Equal(t, builtin.Bool(false), gtResult)

	// Test BuiltinLt (less than)
	ltResult, err := lang.BuiltinLt(builtin.Int64(3), builtin.Int64(7))
	require.NoError(t, err)
	assert.Equal(t, builtin.Bool(true), ltResult)

	ltResult, err = lang.BuiltinLt(builtin.Int64(7), builtin.Int64(3))
	require.NoError(t, err)
	assert.Equal(t, builtin.Bool(false), ltResult)

	// Test BuiltinGte (greater than or equal)
	gteResult, err := lang.BuiltinGte(builtin.Int64(10), builtin.Int64(10))
	require.NoError(t, err)
	assert.Equal(t, builtin.Bool(true), gteResult)

	gteResult, err = lang.BuiltinGte(builtin.Int64(5), builtin.Int64(10))
	require.NoError(t, err)
	assert.Equal(t, builtin.Bool(false), gteResult)

	// Test BuiltinLte (less than or equal)
	lteResult, err := lang.BuiltinLte(builtin.Int64(5), builtin.Int64(5))
	require.NoError(t, err)
	assert.Equal(t, builtin.Bool(true), lteResult)

	lteResult, err = lang.BuiltinLte(builtin.Int64(10), builtin.Int64(5))
	require.NoError(t, err)
	assert.Equal(t, builtin.Bool(false), lteResult)
}

// TestBuiltinPrintln tests the BuiltinPrintln function
func TestBuiltinPrintln(t *testing.T) {
	t.Parallel()
	// Test BuiltinPrintln (this is a placeholder that just returns the string representation)
	result, err := lang.BuiltinPrintln(builtin.String("Hello, World!"))
	require.NoError(t, err)
	assert.Equal(t, builtin.String("\"Hello, World!\""), result)
}

// TestBuiltinShellInfo tests the BuiltinShellInfo function
func TestBuiltinShellInfo(t *testing.T) {
	t.Parallel()
	// Test BuiltinShellInfo
	result, err := lang.BuiltinShellInfo()
	require.NoError(t, err)

	// Should return a Map with shell information
	info, ok := result.(lang.Map)
	require.True(t, ok)

	// Check that it has the expected keys
	name, ok := info.Get(builtin.Keyword("name"))
	require.True(t, ok)
	assert.Equal(t, builtin.String("Wetware Shell"), name)

	version, ok := info.Get(builtin.Keyword("version"))
	require.True(t, ok)
	assert.Equal(t, builtin.String("0.1.0"), version)
}

// TestBuiltinNamespace tests the BuiltinNamespace function
func TestBuiltinNamespace(t *testing.T) {
	t.Parallel()
	// Test BuiltinNamespace
	result, err := lang.BuiltinNamespace()
	require.NoError(t, err)

	// Should return a Map with namespace information
	namespace, ok := result.(lang.Map)
	require.True(t, ok)

	// Check that it has some keys (the default environment)
	keys, err := lang.BuiltinKeys(namespace)
	require.NoError(t, err)

	keyList, ok := keys.(core.Seq)
	require.True(t, ok)

	count, err := keyList.Count()
	require.NoError(t, err)
	assert.Greater(t, count, 0) // Should have some keys
}

// TestBuiltinKeys tests the BuiltinKeys function
func TestBuiltinKeys(t *testing.T) {
	t.Parallel()
	// Create a test map
	testMap := lang.Map{
		builtin.Keyword("key1"): builtin.String("value1"),
		builtin.Keyword("key2"): builtin.Int64(42),
	}

	// Test BuiltinKeys
	result, err := lang.BuiltinKeys(testMap)
	require.NoError(t, err)

	// Should return a list of keys
	keys, ok := result.(core.Seq)
	require.True(t, ok)

	count, err := keys.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count) // Should have exactly 2 keys
}

// TestPathFunctions tests the path-related functions
func TestPathFunctions(t *testing.T) {
	t.Parallel()
	// Test NewUnixPath with valid path (using a real CID)
	path, err := lang.NewUnixPath("/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa")
	require.NoError(t, err)
	assert.Equal(t, "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa", path.String())

	// Test ToBuiltinString
	builtinStr := path.ToBuiltinString()
	assert.Equal(t, builtin.String("/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa"), builtinStr)
}

// TestPathValidation tests the path validation function
func TestPathValidation(t *testing.T) {
	t.Parallel()
	// Test valid paths
	validPath, err := lang.NewUnixPath("/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa")
	require.NoError(t, err)
	assert.Equal(t, "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa", validPath.String())

	// Test invalid paths
	_, err = lang.NewUnixPath("/invalid/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Unix path")

	_, err = lang.NewUnixPath("no-slash")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Unix path")

	_, err = lang.NewUnixPath("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Unix path")
}

// TestMapFunctions tests the map-related functions
func TestMapFunctions(t *testing.T) {
	t.Parallel()
	// Create a test map
	testMap := lang.Map{
		builtin.Keyword("key1"): builtin.String("value1"),
		builtin.Keyword("key2"): builtin.Int64(42),
		builtin.Keyword("key3"): builtin.Bool(true),
	}

	// Test Get
	value, ok := testMap.Get(builtin.Keyword("key1"))
	require.True(t, ok)
	assert.Equal(t, builtin.String("value1"), value)

	// Test Get with non-existent key
	_, ok = testMap.Get(builtin.Keyword("nonexistent"))
	require.False(t, ok)

	// Test With (adding a new key-value pair)
	newMap := testMap.With(builtin.Keyword("newKey"), builtin.String("newValue"))

	// Verify the new map has the new key
	value, ok = newMap.Get(builtin.Keyword("newKey"))
	require.True(t, ok)
	assert.Equal(t, builtin.String("newValue"), value)

	// Verify the original map is unchanged
	_, ok = testMap.Get(builtin.Keyword("newKey"))
	require.False(t, ok)

	// Test Without (removing a key)
	removedMap := testMap.Without(builtin.Keyword("key1"))

	// Verify the key is removed
	_, ok = removedMap.Get(builtin.Keyword("key1"))
	require.False(t, ok)

	// Verify other keys are still there
	value, ok = removedMap.Get(builtin.Keyword("key2"))
	require.True(t, ok)
	assert.Equal(t, builtin.Int64(42), value)

	// Test Len
	length := testMap.Len()
	assert.Equal(t, 3, length)

	// Test SExpr
	sexpr, err := testMap.SExpr()
	require.NoError(t, err)
	assert.Contains(t, sexpr, "key1")
	assert.Contains(t, sexpr, "value1")
	assert.Contains(t, sexpr, "key2")
	assert.Contains(t, sexpr, "42")
}

// TestIPLDConsCell tests the IPLD cons cell functionality
func TestIPLDConsCell(t *testing.T) {
	t.Parallel()
	mockIPFS := system.IPFS{}

	// Test creating a cons cell with car and cdr
	car := builtin.String("test-value")
	cell, err := lang.NewIPLDConsCell(mockIPFS, car, nil)
	require.NoError(t, err)
	require.NotNil(t, cell)

	// Verify the cons cell properties
	assert.Equal(t, car, cell.Car)
	assert.NotNil(t, cell.Cdr)
	assert.NotEmpty(t, cell.Cdr.String())

	// Test creating a cons cell with a cdr
	cdr := cell.Cdr
	cell2, err := lang.NewIPLDConsCell(mockIPFS, builtin.String("second-value"), cdr)
	require.NoError(t, err)
	require.NotNil(t, cell2)

	// Verify the second cons cell
	assert.Equal(t, builtin.String("second-value"), cell2.Car)
	assert.NotNil(t, cell2.Cdr)
	assert.NotEqual(t, cdr.String(), cell2.Cdr.String()) // Should be different CIDs
}

// TestIPLDConsCellChain tests building a chain of cons cells
func TestIPLDConsCellChain(t *testing.T) {
	t.Parallel()
	mockIPFS := system.IPFS{}

	// Build a chain of cons cells: (cons "c" (cons "b" (cons "a" nil)))
	var currentCell *lang.IPLDConsCell

	// Start with the last cell (a -> nil)
	cellA, err := lang.NewIPLDConsCell(mockIPFS, builtin.String("a"), nil)
	require.NoError(t, err)
	currentCell = cellA

	// Add cell b -> a
	cellB, err := lang.NewIPLDConsCell(mockIPFS, builtin.String("b"), currentCell.Cdr)
	require.NoError(t, err)
	currentCell = cellB

	// Add cell c -> b
	cellC, err := lang.NewIPLDConsCell(mockIPFS, builtin.String("c"), currentCell.Cdr)
	require.NoError(t, err)

	// Verify the chain
	assert.Equal(t, builtin.String("c"), cellC.Car)
	assert.NotNil(t, cellC.Cdr)
	assert.Equal(t, builtin.String("b"), cellB.Car)
	assert.NotNil(t, cellB.Cdr)
	assert.Equal(t, builtin.String("a"), cellA.Car)
	assert.NotNil(t, cellA.Cdr)
}

// TestLispFormEvaluation tests evaluating simple Lisp forms using our IPLD linked list
func TestLispFormEvaluation(t *testing.T) {
	t.Parallel()
	mockIPFS := system.IPFS{}

	// Test 1: Simple function call form: (+ 1 2)
	// Build the form: (cons "+" (cons 1 (cons 2 nil)))
	form, err := lang.NewIPLDLinkedList(mockIPFS,
		builtin.Symbol("+"),
		builtin.Int64(1),
		builtin.Int64(2))
	require.NoError(t, err)

	// Extract the function name (first element)
	function, err := form.First()
	require.NoError(t, err)
	assert.Equal(t, builtin.Symbol("+"), function)

	// Test 2: List construction form: (cons "a" (list "b" "c"))
	// For now, we'll test with simple forms that don't nest IPLDLinkedList objects
	// since IPLD doesn't know how to marshal them directly

	// Build a simple form: (cons "a" "b")
	consForm, err := lang.NewIPLDLinkedList(mockIPFS,
		builtin.Symbol("cons"),
		builtin.String("a"),
		builtin.String("b"))
	require.NoError(t, err)

	// Extract the function name
	consFunction, err := consForm.First()
	require.NoError(t, err)
	assert.Equal(t, builtin.Symbol("cons"), consFunction)

	// Test 3: Simple function call: (first (list 1 2 3))
	// Build the form: (first (list 1 2 3))
	// For now, we'll test with simple forms that don't nest IPLDLinkedList objects

	// Build a simple form: (first "some-list")
	firstForm, err := lang.NewIPLDLinkedList(mockIPFS,
		builtin.Symbol("first"),
		builtin.String("some-list"))
	require.NoError(t, err)

	// Extract the function name
	firstFunction, err := firstForm.First()
	require.NoError(t, err)
	assert.Equal(t, builtin.Symbol("first"), firstFunction)

	// Test 4: Conditional form: (if condition "positive" "negative")
	// Build a simple conditional form: (if "condition" "positive" "negative")
	conditionalForm, err := lang.NewIPLDLinkedList(mockIPFS,
		builtin.Symbol("if"),
		builtin.String("condition"),
		builtin.String("positive"),
		builtin.String("negative"))
	require.NoError(t, err)

	// Extract the function name
	conditionalFunction, err := conditionalForm.First()
	require.NoError(t, err)
	assert.Equal(t, builtin.Symbol("if"), conditionalFunction)

	// Test 5: Lambda form: (lambda params body)
	// Build a simple lambda form: (lambda "params" "body")
	lambdaForm, err := lang.NewIPLDLinkedList(mockIPFS,
		builtin.Symbol("lambda"),
		builtin.String("params"),
		builtin.String("body"))
	require.NoError(t, err)

	// Extract the function name
	lambdaFunction, err := lambdaForm.First()
	require.NoError(t, err)
	assert.Equal(t, builtin.Symbol("lambda"), lambdaFunction)
}

// TestLispFormTraversal tests traversing and extracting parts of Lisp forms
func TestLispFormTraversal(t *testing.T) {
	t.Parallel()
	mockIPFS := system.IPFS{}

	// Build a simple form: (define factorial "lambda-body")
	// This demonstrates how we can extract parts of a Lisp form

	define, err := lang.NewIPLDLinkedList(mockIPFS,
		builtin.Symbol("define"),
		builtin.Symbol("factorial"),
		builtin.String("lambda-body"))
	require.NoError(t, err)

	// Test form traversal
	// Extract the function name (first element)
	function, err := define.First()
	require.NoError(t, err)
	assert.Equal(t, builtin.Symbol("define"), function)

	// Count the arguments
	count, err := define.Count()
	require.NoError(t, err)
	assert.Equal(t, 3, count) // define, factorial, lambda-body

	// Test that we can extract the symbol name
	// In a real evaluator, we'd traverse the form to get the second element
	// For now, we just verify the structure is correct
	assert.NotNil(t, define)
}

// Helper functions for creating test data

// createTestList creates a core.Seq for testing
func createTestList(items ...any) core.Seq {
	// Convert any to core.Any
	coreItems := make([]core.Any, len(items))
	for i, item := range items {
		coreItems[i] = item
	}
	return builtin.NewList(coreItems...)
}

// MockAnalyzer implements core.Analyzer for testing
type MockAnalyzer struct {
	analyzedForms []any
	shouldError   bool
	errorMsg      string
}

// Analyze implements core.Analyzer
func (ma *MockAnalyzer) Analyze(env core.Env, form core.Any) (core.Expr, error) {
	ma.analyzedForms = append(ma.analyzedForms, form)

	if ma.shouldError {
		return nil, errors.New(ma.errorMsg)
	}

	// Return a simple constant expression
	return &MockExpr{value: form}, nil
}

// MockExpr implements core.Expr for testing
type MockExpr struct {
	value any
}

// Eval implements core.Expr
func (me *MockExpr) Eval(env core.Env) (core.Any, error) {
	return me.value, nil
}
