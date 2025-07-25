package lang_test

import (
	"context"
	"testing"

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

	require.Equal(t, "0x746573742064617461", buffer.String(), "Buffer hex representation mismatch")

	t.Logf("Successfully tested IPFSCat with UnixPath argument")
}

// TestBuffer tests the Buffer type directly
func TestBuffer(t *testing.T) {
	// Test empty buffer
	emptyBuffer := &lang.Buffer{}
	require.Equal(t, "0x", emptyBuffer.String(), "Empty buffer should return '0x'")

	// Test buffer with data
	testData := []byte("test data")
	buffer := &lang.Buffer{}
	buffer.Write(testData)
	require.Equal(t, "0x746573742064617461", buffer.String(), "Buffer hex representation mismatch")
}

// TestIPFSAdd tests the IPFSAdd function
func TestIPFSAdd(t *testing.T) {
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}
	mock := system.IPFS_ServerToClient(mockServer)

	// Create the IPFSAdd function
	addFunc := lang.IPFSAdd{IPFS: mock}

	// Create a Buffer with test data
	buffer := &lang.Buffer{}
	testData := []byte("test data for add")
	buffer.Write(testData)

	// Test with Buffer
	result, err := addFunc.Invoke(buffer)
	require.NoError(t, err, "Failed to invoke add with Buffer")

	cid, ok := result.(string)
	require.True(t, ok, "Expected string result, got %T", result)

	require.Equal(t, "QmTest123", cid, "CID mismatch")

	t.Logf("Successfully tested IPFSAdd with Buffer argument")
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
