package util

import (
	"os"
	"reflect"
	"testing"
)

func TestGetFDMap(t *testing.T) {
	// Test with no FD environment variables
	fdMap := GetFDMap()
	if len(fdMap) != 0 {
		t.Errorf("Expected empty map with no FD env vars, got %v", fdMap)
	}

	// Test with FD environment variables
	os.Setenv("WW_FD_DB", "4")
	os.Setenv("WW_FD_CACHE", "5")
	os.Setenv("WW_FD_LOGS", "6")
	os.Setenv("OTHER_VAR", "value") // Should be ignored
	defer func() {
		os.Unsetenv("WW_FD_DB")
		os.Unsetenv("WW_FD_CACHE")
		os.Unsetenv("WW_FD_LOGS")
		os.Unsetenv("OTHER_VAR")
	}()

	fdMap = GetFDMap()
	expected := map[string]int{
		"db":    4,
		"cache": 5,
		"logs":  6,
	}

	if !reflect.DeepEqual(fdMap, expected) {
		t.Errorf("GetFDMap() = %v, want %v", fdMap, expected)
	}
}


