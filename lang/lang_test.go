package lang_test

import (
	"context"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
)

// TestIPFSCat tests the standalone IPFSCat function
func TestIPFSCat(t *testing.T) {
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}
	mock := system.IPFS_ServerToClient(mockServer)

	// Create the IPFSCat function
	catFunc := lang.IPFSCat{IPFS: mock}

	// Test with UnixPath
	unixPath, err := lang.NewUnixPath("/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa")
	require.NoError(t, err, "Failed to create UnixPath")

	result, err := catFunc.Invoke(unixPath)
	require.NoError(t, err, "Failed to invoke cat with UnixPath")

	buffer, ok := result.(*lang.Buffer)
	require.True(t, ok, "Expected *lang.Buffer result, got %T", result)

	require.Equal(t, "0x746573742064617461", buffer.AsHex(), "Buffer hex representation mismatch")

	t.Logf("Successfully tested IPFSCat with UnixPath argument")
}

// TestBuffer tests the Buffer type directly
func TestBuffer(t *testing.T) {
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
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}
	mock := system.IPFS_ServerToClient(mockServer)

	// Create the IPFSAdd function
	addFunc := lang.IPFSAdd{IPFS: mock}

	// Create a Buffer with test data
	testData := []byte("test data for add")
	buffer := &lang.Buffer{Mem: testData}

	// Test with Buffer
	result, err := addFunc.Invoke(buffer)
	require.NoError(t, err, "Failed to invoke add with Buffer")

	cid, ok := result.(string)
	require.True(t, ok, "Expected string result, got %T", result)

	require.Equal(t, "QmTest123", cid, "CID mismatch")

	t.Logf("Successfully tested IPFSAdd with Buffer argument")
}

// TestGoSpecialForm tests the Go special form argument validation
func TestGoSpecialForm(t *testing.T) {
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
	info.SetType("file")
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

	// Extract path
	path, err := params.Path()
	if err == nil {
		m.spawnPath = path
	}

	// Extract arguments
	args, err := params.Args()
	if err == nil {
		m.spawnArgs = make([]string, args.Len())
		for i := 0; i < args.Len(); i++ {
			arg, err := args.At(i)
			if err == nil {
				m.spawnArgs[i] = arg
			}
		}
	}

	// Extract directory
	dir, err := params.Dir()
	if err == nil {
		m.spawnDir = dir
	}

	// Extract environment
	env, err := params.Env()
	if err == nil {
		m.spawnEnv = make([]string, env.Len())
		for i := 0; i < env.Len(); i++ {
			envVar, err := env.At(i)
			if err == nil {
				m.spawnEnv[i] = envVar
			}
		}
	}

	// Allocate results
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	if m.shouldSucceed {
		// Create a successful result with a mock cell
		optionalCell, err := results.NewCell()
		if err != nil {
			return err
		}
		// Create a mock cell server and convert to client
		mockCell := &MockCell{}
		cellClient := system.Cell_ServerToClient(mockCell)
		optionalCell.SetCell(cellClient)
	} else {
		// Create an error result
		optionalCell, err := results.NewCell()
		if err != nil {
			return err
		}
		optionalCell.SetErr()
		errStruct := optionalCell.Err()
		errStruct.SetStatus(1)
		errStruct.SetBody([]byte("mock error"))
	}

	return nil
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
