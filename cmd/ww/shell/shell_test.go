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
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}

	// Create a client from the server
	mock := system.IPFS_ServerToClient(mockServer)

	// Create environment with multiple values
	env := core.New(map[string]core.Any{
		"number":     42,
		"string":     "hello",
		"bool":       true,
		"capability": lang.Session{IPFS: mock},
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
		{"capability", "capability", lang.Session{IPFS: mock}},
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
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}

	// Create a client from the server
	mock := system.IPFS_ServerToClient(mockServer)

	// Create the invokable wrapper
	sess := lang.Session{IPFS: mock}

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
