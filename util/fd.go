package util

import (
	"os"
	"strconv"
	"strings"
)

// GetFDMap returns a map of names to file descriptor numbers from environment variables
// This is a utility function for child processes to easily access their FDs
func GetFDMap() map[string]int {
	fdMap := make(map[string]int)

	for _, env := range os.Environ() {
		if len(env) > 6 && strings.HasPrefix(env, "WW_FD_") {
			// Extract name and FD number from WW_FD_NAME=fdnum
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				name := strings.ToLower(parts[0][6:]) // Remove "WW_FD_" prefix and convert to lowercase
				if fd, err := strconv.Atoi(parts[1]); err == nil {
					fdMap[name] = fd
				}
			}
		}
	}

	return fdMap
}
