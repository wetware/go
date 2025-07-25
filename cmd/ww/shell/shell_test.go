package shell_test

import (
	"context"
	"testing"

	"strings"

	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
)

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
	return results.SetBody([]byte("test data"))
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
	// Mock implementation - return empty list
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	cids, err := results.NewCids(0)
	if err != nil {
		return err
	}
	return results.SetCids(cids)
}

// Id implements system.IPFS_Server.Id
func (m *MockIPFSServer) Id(ctx context.Context, call system.IPFS_id) error {
	// Mock implementation
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	peerInfo, err := results.NewPeerInfo()
	if err != nil {
		return err
	}
	peerInfo.SetId("QmTestPeer")
	return results.SetPeerInfo(peerInfo)
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

// TestShellEnvironment tests that the environment is set up correctly
func TestShellEnvironment(t *testing.T) {
	t.Parallel()
	// Create a simple environment with a test value
	env := core.New(map[string]core.Any{
		"test": 42,
	})

	// Check that test is in the environment
	testValue, err := env.Resolve("test")
	if err != nil {
		t.Fatalf("Failed to resolve test: %v", err)
	}
	if testValue != 42 {
		t.Errorf("Expected 42, got %v", testValue)
	}
}

// TestEnvironmentWithMultipleValues tests environment with multiple values
func TestEnvironmentWithMultipleValues(t *testing.T) {
	t.Parallel()
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}

	// Create a client from the server
	mock := system.IPFS_ServerToClient(mockServer)

	// Create environment with multiple values
	env := core.New(map[string]core.Any{
		"number":     42,
		"string":     "hello",
		"bool":       true,
		"capability": lang.IPFSCat{IPFS: mock},
	})

	// Test resolving each value
	testCases := []struct {
		name     string
		key      string
		expected interface{}
	}{
		{"number", "number", 42},
		{"string", "string", "hello"},
		{"bool", "bool", true},
		{"capability", "capability", lang.IPFSCat{IPFS: mock}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := env.Resolve(tc.key)
			if err != nil {
				t.Fatalf("Failed to resolve %s: %v", tc.key, err)
			}
			if value != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, value)
			}
		})
	}
}

// TestInvokableWithMockIPFS tests the Invokable wrapper with mock IPFS
func TestInvokableWithMockIPFS(t *testing.T) {
	t.Parallel()
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}

	// Create a client from the server
	mock := system.IPFS_ServerToClient(mockServer)

	// Create the invokable wrapper
	sess := lang.IPFSCat{IPFS: mock}

	// Test that we can access the client
	if sess.IPFS != mock {
		t.Errorf("Expected client to be the mock, got %v", sess.IPFS)
	}

	// Test that we can return the session itself when no arguments are provided
	result, err := sess.Invoke()
	if err != nil {
		t.Fatalf("Failed to invoke with no arguments: %v", err)
	}

	// Should return the session itself
	require.Equal(t, sess, result, "identity law not verified:  `(session)` should return `session`")

	t.Logf("Successfully created mock IPFS capability wrapped in Invokable")
	t.Logf("Mock server test value: %d", mockServer.testValue)
}

// TestEnvironmentNotFound tests resolving non-existent values
func TestEnvironmentNotFound(t *testing.T) {
	t.Parallel()
	// Create a simple environment
	env := core.New(map[string]core.Any{
		"test": 42,
	})

	// Try to resolve a non-existent value
	_, err := env.Resolve("nonexistent")
	if err == nil {
		t.Error("Expected error when resolving non-existent value")
	}
}

func TestUnixPathReader(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid ipfs path",
			input:   "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			want:    "/ipfs/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			wantErr: false,
		},
		{
			name:    "valid ipld path",
			input:   "/ipld/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			want:    "/ipld/QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa",
			wantErr: false,
		},
		{
			name:    "invalid path starting with slash",
			input:   "/invalid/path",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader with our custom macro
			rd := reader.New(strings.NewReader(tt.input))
			rd.SetMacro('/', false, lang.UnixPathReader())

			// Read one form
			result, err := rd.One()

			if tt.wantErr {
				if err == nil {
					t.Errorf("unixPathReader() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unixPathReader() unexpected error: %v", err)
				return
			}

			// Check if result is a UnixPath
			unixPath, ok := result.(*lang.UnixPath)
			if !ok {
				t.Errorf("unixPathReader() returned %T, want *lang.UnixPath", result)
				return
			}

			if unixPath.String() != tt.want {
				t.Errorf("unixPathReader() = %v, want %v", unixPath.String(), tt.want)
			}
		})
	}
}

// MockTerminalSession implements auth.Terminal_login_Results for testing
type MockTerminalSession struct {
	hasConsole bool
	hasIPFS    bool
	hasExec    bool
	console    system.Console
	ipfs       system.IPFS
	exec       system.Executor
}

func (m *MockTerminalSession) HasConsole() bool { return m.hasConsole }
func (m *MockTerminalSession) HasIpfs() bool    { return m.hasIPFS }
func (m *MockTerminalSession) HasExec() bool    { return m.hasExec }
func (m *MockTerminalSession) Console() system.Console {
	if !m.hasConsole {
		return system.Console{}
	}
	return m.console
}
func (m *MockTerminalSession) Ipfs() system.IPFS {
	if !m.hasIPFS {
		return system.IPFS{}
	}
	return m.ipfs
}
func (m *MockTerminalSession) Exec() system.Executor {
	if !m.hasExec {
		return system.Executor{}
	}
	return m.exec
}

// TestCapabilityWithholding tests that capabilities are properly withheld when not granted
func TestCapabilityWithholding(t *testing.T) {
	t.Parallel()

	// Create mock capabilities
	mockConsole := system.Console_ServerToClient(&MockConsoleServer{})
	mockIPFS := system.IPFS_ServerToClient(&MockIPFSServer{})
	mockExec := system.Executor_ServerToClient(&MockExecutorServer{})

	testCases := []struct {
		name           string
		hasConsole     bool
		hasIPFS        bool
		hasExec        bool
		expectedKeys   []string
		unexpectedKeys []string
	}{
		{
			name:           "no capabilities granted",
			hasConsole:     false,
			hasIPFS:        false,
			hasExec:        false,
			expectedKeys:   []string{},
			unexpectedKeys: []string{"println", "cat", "add", "ls", "stat", "pin", "unpin", "pins", "id", "connect", "peers", "go"},
		},
		{
			name:           "only console granted",
			hasConsole:     true,
			hasIPFS:        false,
			hasExec:        false,
			expectedKeys:   []string{"println"},
			unexpectedKeys: []string{"cat", "add", "ls", "stat", "pin", "unpin", "pins", "id", "connect", "peers", "go"},
		},
		{
			name:           "only IPFS granted",
			hasConsole:     false,
			hasIPFS:        true,
			hasExec:        false,
			expectedKeys:   []string{"cat", "add", "ls", "stat", "pin", "unpin", "pins", "id", "connect", "peers"},
			unexpectedKeys: []string{"println", "go"},
		},
		{
			name:           "only exec granted",
			hasConsole:     false,
			hasIPFS:        false,
			hasExec:        true,
			expectedKeys:   []string{"go"},
			unexpectedKeys: []string{"println", "cat", "add", "ls", "stat", "pin", "unpin", "pins", "id", "connect", "peers"},
		},
		{
			name:           "all capabilities granted",
			hasConsole:     true,
			hasIPFS:        true,
			hasExec:        true,
			expectedKeys:   []string{"println", "cat", "add", "ls", "stat", "pin", "unpin", "pins", "id", "connect", "peers", "go"},
			unexpectedKeys: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock session with specified capabilities
			sess := &MockTerminalSession{
				hasConsole: tc.hasConsole,
				hasIPFS:    tc.hasIPFS,
				hasExec:    tc.hasExec,
				console:    mockConsole,
				ipfs:       mockIPFS,
				exec:       mockExec,
			}

			// Create environment using the cell function logic
			globals := make(map[string]core.Any)

			// Conditionally add console capability
			if sess.HasConsole() {
				console := sess.Console()
				globals["println"] = lang.ConsolePrintln{Console: console}
			}

			// Conditionally add IPFS capabilities
			if sess.HasIpfs() {
				ipfs := sess.Ipfs()
				globals["cat"] = lang.IPFSCat{IPFS: ipfs}
				globals["add"] = lang.IPFSAdd{IPFS: ipfs}
				globals["ls"] = &lang.IPFSLs{IPFS: ipfs}
				globals["stat"] = &lang.IPFSStat{IPFS: ipfs}
				globals["pin"] = &lang.IPFSPin{IPFS: ipfs}
				globals["unpin"] = &lang.IPFSUnpin{IPFS: ipfs}
				globals["pins"] = &lang.IPFSPins{IPFS: ipfs}
				globals["id"] = &lang.IPFSId{IPFS: ipfs}
				globals["connect"] = &lang.IPFSConnect{IPFS: ipfs}
				globals["peers"] = &lang.IPFSPeers{IPFS: ipfs}
			}

			// Conditionally add process execution capability
			if sess.HasExec() {
				exec := sess.Exec()
				globals["go"] = lang.Go{Executor: exec}
			}

			env := core.New(globals)

			// Test that expected keys are present
			for _, key := range tc.expectedKeys {
				value, err := env.Resolve(key)
				if err != nil {
					t.Errorf("Expected key '%s' to be present, but got error: %v", key, err)
				}
				if value == nil {
					t.Errorf("Expected key '%s' to have a non-nil value", key)
				}
			}

			// Test that unexpected keys are not present
			for _, key := range tc.unexpectedKeys {
				_, err := env.Resolve(key)
				if err == nil {
					t.Errorf("Expected key '%s' to be absent, but it was found", key)
				}
			}
		})
	}
}

// MockConsoleServer implements system.Console_Server for testing
type MockConsoleServer struct{}

func (m *MockConsoleServer) Println(ctx context.Context, call system.Console_println) error {
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	results.SetN(10) // Return 10 bytes written
	return nil
}

// MockCellServer implements system.Cell_Server for testing
type MockCellServer struct{}

func (m *MockCellServer) Wait(ctx context.Context, call system.Cell_wait) error {
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	// Return a mock result
	result, err := results.NewResult()
	if err != nil {
		return err
	}
	result.SetOk()
	return results.SetResult(result)
}

// MockExecutorServer implements system.Executor_Server for testing
type MockExecutorServer struct{}

func (m *MockExecutorServer) Spawn(ctx context.Context, call system.Executor_spawn) error {
	results, err := call.AllocResults()
	if err != nil {
		return err
	}
	// Return a mock cell
	optionalCell, err := results.NewCell()
	if err != nil {
		return err
	}
	// Create a mock Cell and set it
	mockCell := system.Cell_ServerToClient(&MockCellServer{})
	optionalCell.SetCell(mockCell)
	return results.SetCell(optionalCell)
}
