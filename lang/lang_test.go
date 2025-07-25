package lang_test

import (
	"context"
	"testing"

	"github.com/spy16/slurp/core"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
)

var _ core.Invokable = (*lang.Session)(nil)

func TestIPFSWrapper(t *testing.T) {
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}

	// Create the IPFS wrapper
	mock := system.IPFS_ServerToClient(mockServer)
	wrapper := lang.Session{IPFS: mock}

	// Test Cat method
	data, err := wrapper.Cat("QmTest123")
	if err != nil {
		t.Fatalf("Failed to call Cat: %v", err)
	}
	t.Logf("Direct Cat call returned: '%s' (length: %d)", string(data), len(data))
	if string(data) != "test data" {
		t.Errorf("Expected 'test data', got '%s'", string(data))
	}

	// Test Add method
	cid, err := wrapper.Add([]byte("test data"))
	if err != nil {
		t.Fatalf("Failed to call Add: %v", err)
	}
	if cid != "QmTest123" {
		t.Errorf("Expected 'QmTest123', got '%s'", cid)
	}

	// Test Pins method
	pins, err := wrapper.Pins()
	if err != nil {
		t.Fatalf("Failed to call Pins: %v", err)
	}
	if len(pins) != 2 {
		t.Errorf("Expected 2 pins, got %d", len(pins))
	}
	if pins[0] != "QmTest1" || pins[1] != "QmTest2" {
		t.Errorf("Expected ['QmTest1', 'QmTest2'], got %v", pins)
	}

	// Test Invoke method with Cat
	result, err := wrapper.Invoke("Cat", "QmTest123")
	if err != nil {
		t.Fatalf("Failed to invoke Cat: %v", err)
	}
	data, ok := result.([]byte)
	if !ok {
		t.Fatalf("Expected []byte result, got %T", result)
	}
	t.Logf("Invoke Cat call returned: '%s' (length: %d)", string(data), len(data))
	if string(data) != "test data" {
		t.Errorf("Expected 'test data', got '%s'", string(data))
	}

	// Test Invoke method with no arguments
	result, err = wrapper.Invoke()
	if err != nil {
		t.Fatalf("Failed to invoke with no arguments: %v", err)
	}
	if result != wrapper {
		t.Errorf("Expected wrapper to return itself, got %v", result)
	}

	t.Logf("Successfully tested IPFSWrapper with mock server")
}

// TestDotNotation tests the dot notation syntax like (ipfs.Cat "QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa")
func TestDotNotation(t *testing.T) {
	// Create a mock IPFS server
	mockServer := &MockIPFSServer{testValue: 42}

	// Create the IPFS wrapper
	mock := system.IPFS_ServerToClient(mockServer)
	session := lang.Session{IPFS: mock}

	// Create an environment with the ipfs session
	env := core.New(map[string]core.Any{
		"ipfs": session,
	})

	// Test resolving ipfs from the environment
	ipfsValue, err := env.Resolve("ipfs")
	if err != nil {
		t.Fatalf("Failed to resolve ipfs: %v", err)
	}

	// ipfs should resolve to our session
	ipfsSession, ok := ipfsValue.(lang.Session)
	if !ok {
		t.Fatalf("Expected ipfs to resolve to Session, got %T", ipfsValue)
	}

	// Test resolving Cat method from the session
	catMethod, err := ipfsSession.Resolve("Cat")
	if err != nil {
		t.Fatalf("Failed to resolve Cat method: %v", err)
	}

	// Cat should be invokable
	invokable, ok := catMethod.(core.Invokable)
	if !ok {
		t.Fatalf("Expected Cat to be invokable, got %T", catMethod)
	}

	// Test invoking Cat with a CID
	result, err := invokable.Invoke("QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa")
	if err != nil {
		t.Fatalf("Failed to invoke Cat: %v", err)
	}

	// Result should be the data from IPFS
	data, ok := result.([]byte)
	if !ok {
		t.Fatalf("Expected []byte result, got %T", result)
	}

	t.Logf("Dot notation test returned: '%s' (length: %d)", string(data), len(data))
	if string(data) != "test data" {
		t.Errorf("Expected 'test data', got '%s'", string(data))
	}

	t.Logf("Successfully tested dot notation: (ipfs.Cat \"QmYJKWYVWwJmJpK4N1vRNcZ9uVQYfLRXU9uK9kfiMWQuoa\")")
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
