package run

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFDManager(t *testing.T) {
	t.Parallel()
	fm := NewFDManager()
	assert.NotNil(t, fm, "NewFDManager returned nil")

	assert.Equal(t, 0, len(fm.configs), "expected empty configs, got %d", len(fm.configs))
}

func TestParseFDFlag(t *testing.T) {
	t.Parallel()
	fm := NewFDManager()

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
			name:    "duplicate name",
			value:   "db=4",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fm.ParseFDFlag(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrepareFDs(t *testing.T) {
	t.Parallel()
	fm := NewFDManager()

	// Add test configs
	err := fm.ParseFDFlag("db=3")
	require.NoError(t, err, "failed to add config: %v", err)

	err = fm.ParseFDFlag("cache=4")
	require.NoError(t, err, "failed to add config: %v", err)

	// Prepare fds
	files, err := fm.PrepareFDs()
	require.NoError(t, err, "PrepareFDs() error = %v", err)

	assert.Equal(t, 2, len(files), "expected 2 files, got %d", len(files))

	// Check that target fds were auto-assigned (order may vary due to map iteration)
	config1, exists := fm.configs["db"]
	assert.True(t, exists, "db config not found")

	config2, exists := fm.configs["cache"]
	assert.True(t, exists, "cache config not found")

	// Both should have target FDs assigned, but order may vary
	assert.True(t, config1.TargetFD >= 3 && config1.TargetFD <= 4, "expected target fd 3 or 4, got %d", config1.TargetFD)
	assert.True(t, config2.TargetFD >= 3 && config2.TargetFD <= 4, "expected target fd 3 or 4, got %d", config2.TargetFD)
	assert.NotEqual(t, config1.TargetFD, config2.TargetFD, "expected different target fds, both got %d", config1.TargetFD)
}

func TestGenerateEnvVars(t *testing.T) {
	t.Parallel()
	fm := NewFDManager()

	// Add test configs
	err := fm.ParseFDFlag("db=3")
	require.NoError(t, err, "failed to add config: %v", err)

	err = fm.ParseFDFlag("cache=4")
	require.NoError(t, err, "failed to add config: %v", err)

	// Prepare fds to set target fds
	_, err = fm.PrepareFDs()
	require.NoError(t, err, "failed to prepare fds: %v", err)

	// Generate environment variables
	envVars := fm.GenerateEnvVars()

	// Check for expected environment variables (order may vary)
	expectedNames := []string{"WW_FD_DB", "WW_FD_CACHE"}

	for _, name := range expectedNames {
		found := false
		for _, envVar := range envVars {
			if strings.HasPrefix(envVar, name+"=") {
				found = true
				// Verify the value is a valid fd number (3 or 4)
				value := strings.TrimPrefix(envVar, name+"=")
				assert.True(t, value == "3" || value == "4", "expected %s to be 3 or 4, got %s", name, value)
				break
			}
		}
		assert.True(t, found, "expected environment variable %s not found", name)
	}
}

func TestClose(t *testing.T) {
	t.Parallel()
	fm := NewFDManager()

	// Add test configs
	err := fm.ParseFDFlag("db=3")
	require.NoError(t, err, "failed to add config: %v", err)

	// Prepare fds to create files
	_, err = fm.PrepareFDs()
	require.NoError(t, err, "failed to prepare fds: %v", err)

	// Close all fds
	err = fm.Close()
	assert.NoError(t, err, "Close() error = %v", err)
}
