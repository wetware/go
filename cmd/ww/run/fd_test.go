package run

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestParseFDFlag(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantFD   int
		wantErr  bool
	}{
		{"db=3", "db", 3, false},
		{"cache=4", "cache", 4, false},
		{"input=0", "input", 0, false},
		{"invalid", "", 0, true},
		{"=3", "", 0, true},
		{"db=", "", 0, true},
		{"db=-1", "", 0, true},
		{"db=abc", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, fd, err := ParseFDFlag(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFDFlag() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFDFlag() unexpected error: %v", err)
				return
			}

			if name != tt.wantName {
				t.Errorf("ParseFDFlag() name = %v, want %v", name, tt.wantName)
			}

			if fd != tt.wantFD {
				t.Errorf("ParseFDFlag() fd = %v, want %v", fd, tt.wantFD)
			}
		})
	}
}

func TestNewFDManager(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		wantErr bool
	}{
		{"valid single", []string{"db=3"}, false},
		{"valid multiple", []string{"db=3", "cache=4"}, false},
		{"duplicate names", []string{"db=3", "db=4"}, true},
		{"invalid format", []string{"invalid"}, true},
		{"empty list", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, err := NewFDManager(tt.flags)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewFDManager() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewFDManager() unexpected error: %v", err)
				return
			}

			if fm == nil {
				t.Errorf("NewFDManager() returned nil manager")
			}
		})
	}
}

func TestFDManager_GenerateEnvVars(t *testing.T) {
	fm, err := NewFDManager([]string{"db=3", "cache=4"})
	if err != nil {
		t.Fatalf("Failed to create FDManager: %v", err)
	}

	envVars := fm.GenerateEnvVars()
	expected := []string{"WW_FD_DB=3", "WW_FD_CACHE=4"}

	if !reflect.DeepEqual(envVars, expected) {
		t.Errorf("GenerateEnvVars() = %v, want %v", envVars, expected)
	}
}

func TestFDManager_Close(t *testing.T) {
	// Create a temporary file to get a valid FD
	tmpFile, err := os.CreateTemp("", "fd-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Get the FD number
	fd := int(tmpFile.Fd())

	fm, err := NewFDManager([]string{fmt.Sprintf("test=%d", fd)})
	if err != nil {
		t.Fatalf("Failed to create FDManager: %v", err)
	}

	// Close should not error
	if err := fm.Close(); err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}
}
