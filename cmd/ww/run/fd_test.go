package run

import (
	"os"
	"strings"
	"testing"
)

func TestNewFDManager(t *testing.T) {
	t.Parallel()
	fm := NewFDManager(false)
	if fm == nil {
		t.Fatal("NewFDManager returned nil")
	}

	if len(fm.configs) != 0 {
		t.Errorf("expected empty configs, got %d", len(fm.configs))
	}

	if fm.verbose != false {
		t.Errorf("expected verbose=false, got %t", fm.verbose)
	}
}

func TestParseFDFlag(t *testing.T) {
	t.Parallel()
	fm := NewFDManager(false)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid basic fd",
			value:   "db=3",
			wantErr: false,
		},
		{
			name:    "valid fd with options",
			value:   "cache=4,mode=rw,type=socket",
			wantErr: false,
		},
		{
			name:    "invalid format - missing fdnum",
			value:   "db",
			wantErr: true,
		},
		{
			name:    "invalid format - missing equals",
			value:   "db3",
			wantErr: true,
		},
		{
			name:    "invalid fd number",
			value:   "db=abc",
			wantErr: true,
		},
		{
			name:    "invalid mode",
			value:   "db=3,mode=invalid",
			wantErr: true,
		},
		{
			name:    "invalid type",
			value:   "db=3,type=invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fm.ParseFDFlag(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFDFlag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseFDMapFlag(t *testing.T) {
	t.Parallel()
	fm := NewFDManager(false)

	// First add a config
	err := fm.ParseFDFlag("db=3")
	if err != nil {
		t.Fatalf("failed to add initial config: %v", err)
	}

	// Test valid mapping
	err = fm.ParseFDMapFlag("db=10")
	if err != nil {
		t.Errorf("ParseFDMapFlag() error = %v", err)
	}

	config, exists := fm.configs["db"]
	if !exists {
		t.Fatal("config not found")
	}

	if config.TargetFD != 10 {
		t.Errorf("expected TargetFD=10, got %d", config.TargetFD)
	}

	// Test invalid mapping - name not found
	err = fm.ParseFDMapFlag("nonexistent=5")
	if err == nil {
		t.Error("expected error for nonexistent name")
	}

	// Test invalid format
	err = fm.ParseFDMapFlag("invalid-format")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestParseFDCTLFlag(t *testing.T) {
	t.Parallel()
	fm := NewFDManager(false)

	// Test inherit format
	err := fm.ParseFDCTLFlag("inherit:5")
	if err != nil {
		t.Errorf("ParseFDCTLFlag(inherit:5) error = %v", err)
	}

	config, exists := fm.configs["inherit_5"]
	if !exists {
		t.Fatal("inherit config not found")
	}

	if config.SourceFD != 5 {
		t.Errorf("expected SourceFD=5, got %d", config.SourceFD)
	}

	// Test unix socket path (should fail for now)
	err = fm.ParseFDCTLFlag("/path/to/socket")
	if err == nil {
		t.Error("expected error for unix socket path")
	}
}

func TestUseSystemdFDs(t *testing.T) {
	t.Parallel()
	fm := NewFDManager(false)

	// Test without systemd environment
	err := fm.UseSystemdFDs("listen")
	if err == nil {
		t.Error("expected error when systemd not available")
	}

	// Test with systemd environment
	os.Setenv("LISTEN_FDS", "2")
	os.Setenv("LISTEN_PID", "12345")

	err = fm.UseSystemdFDs("listen")
	if err == nil {
		t.Error("expected error for PID mismatch")
	}

	// Test with correct PID
	os.Setenv("LISTEN_PID", "1") // Use a PID that won't match
	err = fm.UseSystemdFDs("listen")
	if err == nil {
		t.Error("expected error for PID mismatch")
	}

	// Clean up
	os.Unsetenv("LISTEN_FDS")
	os.Unsetenv("LISTEN_PID")
}

func TestParseFDFromFile(t *testing.T) {
	t.Parallel()
	fm := NewFDManager(false)

	// Test with invalid file path
	err := fm.ParseFDFromFile("/nonexistent/file")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}

	// Test with valid capability table content
	content := `
	;; Cap table format
	(
		(fd :name "stdin" :fd 0 :mode "r" :target 0)
		(fd :name "stdout" :fd 1 :mode "w" :target 1)
		(fd :name "logs" :fd 5 :mode "w" :type "file" :target 9)
	)`

	// Create temporary file
	tmpfile, err := os.CreateTemp("", "imports_*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write content to file
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	// Parse the file
	err = fm.ParseFDFromFile(tmpfile.Name())
	if err != nil {
		t.Errorf("ParseFDFromFile() error = %v", err)
	}

	// Check that configs were created
	expectedNames := []string{"stdin", "stdout", "logs"}
	for _, name := range expectedNames {
		if _, exists := fm.configs[name]; !exists {
			t.Errorf("expected config for %s not found", name)
		}
	}
}

func TestPrepareFDs(t *testing.T) {
	t.Parallel()
	fm := NewFDManager(false)

	// Add some test configs
	err := fm.ParseFDFlag("db=3")
	if err != nil {
		t.Fatalf("failed to add config: %v", err)
	}

	err = fm.ParseFDFlag("cache=4")
	if err != nil {
		t.Fatalf("failed to add config: %v", err)
	}

	// Test duplicate target fd - this should be caught in PrepareFDs
	err = fm.ParseFDMapFlag("db=10")
	if err != nil {
		t.Fatalf("failed to set target fd: %v", err)
	}

	err = fm.ParseFDMapFlag("cache=10")
	if err != nil {
		t.Fatalf("failed to set target fd: %v", err)
	}

	// Now try to prepare fds - this should fail due to duplicate target
	_, err = fm.PrepareFDs()
	if err == nil {
		t.Error("expected error for duplicate target fd")
	}

	// Test auto-assignment with different target fd
	err = fm.ParseFDMapFlag("cache=11")
	if err != nil {
		t.Fatalf("failed to set target fd: %v", err)
	}

	// Now prepare fds should succeed
	files, err := fm.PrepareFDs()
	if err != nil {
		t.Errorf("PrepareFDs() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestGenerateEnvVars(t *testing.T) {
	t.Parallel()
	fm := NewFDManager(false)

	// Add test configs
	err := fm.ParseFDFlag("db=3")
	if err != nil {
		t.Fatalf("failed to add config: %v", err)
	}

	err = fm.ParseFDMapFlag("db=10")
	if err != nil {
		t.Fatalf("failed to set target fd: %v", err)
	}

	// Generate environment variables
	envVars := fm.GenerateEnvVars()

	// Check WW_FDS
	var wwFDS string
	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, "WW_FDS=") {
			wwFDS = strings.TrimPrefix(envVar, "WW_FDS=")
			break
		}
	}

	if wwFDS == "" {
		t.Error("WW_FDS environment variable not found")
	}

	// Check WW_FD_DB
	var wwFdDB string
	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, "WW_FD_DB=") {
			wwFdDB = strings.TrimPrefix(envVar, "WW_FD_DB=")
			break
		}
	}

	if wwFdDB != "10" {
		t.Errorf("expected WW_FD_DB=10, got %s", wwFdDB)
	}
}

func TestCreateSymlinks(t *testing.T) {
	t.Parallel()
	fm := NewFDManager(false)

	// Add config with pathlink
	err := fm.ParseFDFlag("db=3,pathlink=true,target=db.sock")
	if err != nil {
		t.Fatalf("failed to add config: %v", err)
	}

	err = fm.ParseFDMapFlag("db=10")
	if err != nil {
		t.Fatalf("failed to set target fd: %v", err)
	}

	// Create temporary directory
	tmpdir, err := os.MkdirTemp("", "jail_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// Create symlinks
	err = fm.CreateSymlinks(tmpdir)
	if err != nil {
		t.Errorf("CreateSymlinks() error = %v", err)
	}

	// Check that symlink was created
	linkPath := tmpdir + "/db.sock"
	if _, err := os.Lstat(linkPath); err != nil {
		t.Errorf("symlink not created: %v", err)
	}
}

func TestClose(t *testing.T) {
	t.Parallel()
	fm := NewFDManager(false)

	// Add some configs and prepare fds
	err := fm.ParseFDFlag("db=3")
	if err != nil {
		t.Fatalf("failed to add config: %v", err)
	}

	_, err = fm.PrepareFDs()
	if err != nil {
		t.Fatalf("failed to prepare fds: %v", err)
	}

	// Close should not error
	err = fm.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
