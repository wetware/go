package run

import (
	"testing"

	"github.com/ipfs/boxo/path"
)

func TestParseWithChildFlag(t *testing.T) {
	cm := NewChildManager()

	// Test IPFS path (using a valid CID format)
	err := cm.ParseWithChildFlag("db=/ipfs/QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG/db-service")
	if err != nil {
		t.Errorf("failed to parse IPFS path: %v", err)
	}

	// Test local path
	err = cm.ParseWithChildFlag("cache=./local/cache-service")
	if err != nil {
		t.Errorf("failed to parse local path: %v", err)
	}

	// Test FD number
	err = cm.ParseWithChildFlag("storage=5")
	if err != nil {
		t.Errorf("failed to parse FD number: %v", err)
	}

	// Test invalid FD (stdio)
	err = cm.ParseWithChildFlag("invalid=0")
	if err == nil {
		t.Error("expected error for stdio FD 0")
	}

	// Test duplicate name
	err = cm.ParseWithChildFlag("db=/ipfs/QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG/another-service")
	if err == nil {
		t.Error("expected error for duplicate name")
	}

	// Test invalid format
	err = cm.ParseWithChildFlag("invalid-format")
	if err == nil {
		t.Error("expected error for invalid format")
	}

	// Test empty name
	err = cm.ParseWithChildFlag("=./service")
	if err == nil {
		t.Error("expected error for empty name")
	}

	// Verify we have the expected children
	children := cm.GetChildren()
	if len(children) != 3 {
		t.Errorf("expected 3 children, got %d", len(children))
	}

	// Verify first child is IPFS source
	if children[0].Type != IPFSSource {
		t.Errorf("expected first child to be IPFS source, got %v", children[0].Type)
	}

	// Verify second child is local source
	if children[1].Type != LocalSource {
		t.Errorf("expected second child to be local source, got %v", children[1].Type)
	}

	// Verify third child is FD source
	if children[2].Type != FDSource {
		t.Errorf("expected third child to be FD source, got %v", children[2].Type)
	}
}

func TestChildSpecValidation(t *testing.T) {
	// Test IPFS path validation
	ipfsPath, err := path.NewPath("/ipfs/QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG/service")
	if err != nil {
		t.Fatalf("failed to create IPFS path: %v", err)
	}

	child := &ChildSpec{
		Name:   "test",
		Source: ipfsPath,
		Type:   IPFSSource,
	}

	if child.Name != "test" {
		t.Errorf("expected name 'test', got %s", child.Name)
	}

	if child.Type != IPFSSource {
		t.Errorf("expected type IPFSSource, got %v", child.Type)
	}
}
