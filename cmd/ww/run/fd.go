package run

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// FDManager handles file descriptor mapping for child processes
type FDManager struct {
	mappings []fdMapping
}

type fdMapping struct {
	name     string
	sourceFD int
	targetFD int
}

// ParseFDFlag parses a --with-fd flag value in "name=fdnum" format
func ParseFDFlag(value string) (string, int, error) {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid format: expected 'name=fdnum', got '%s'", value)
	}

	name := parts[0]
	if name == "" {
		return "", 0, fmt.Errorf("name cannot be empty")
	}

	fdnum, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid fd number '%s': %v", parts[1], err)
	}

	if fdnum < 0 {
		return "", 0, fmt.Errorf("fd number must be non-negative, got %d", fdnum)
	}

	return name, fdnum, nil
}

// NewFDManager creates a new FD manager from a list of --with-fd flag values
func NewFDManager(fdFlags []string) (*FDManager, error) {
	fm := &FDManager{}

	// Track used names to prevent duplicates
	usedNames := make(map[string]bool)

	for i, flag := range fdFlags {
		name, sourceFD, err := ParseFDFlag(flag)
		if err != nil {
			return nil, fmt.Errorf("flag %d '%s': %w", i+1, flag, err)
		}

		if usedNames[name] {
			return nil, fmt.Errorf("duplicate name '%s' in --with-fd flags", name)
		}

		// Target FD starts at 3 and increments sequentially
		targetFD := 3 + i

		fm.mappings = append(fm.mappings, fdMapping{
			name:     name,
			sourceFD: sourceFD,
			targetFD: targetFD,
		})

		usedNames[name] = true
	}

	return fm, nil
}

// PrepareFDs duplicates source file descriptors and returns them for ExtraFiles
func (fm *FDManager) PrepareFDs() ([]*os.File, error) {
	var files []*os.File

	for _, mapping := range fm.mappings {
		// Duplicate the source FD to avoid conflicts
		newFD, err := syscall.Dup(mapping.sourceFD)
		if err != nil {
			return nil, fmt.Errorf("failed to duplicate fd %d for '%s': %w", mapping.sourceFD, mapping.name, err)
		}

		// Create os.File from the duplicated FD
		file := os.NewFile(uintptr(newFD), mapping.name)
		if file == nil {
			syscall.Close(newFD)
			return nil, fmt.Errorf("failed to create os.File for fd %d", newFD)
		}

		files = append(files, file)
	}

	return files, nil
}

// GenerateEnvVars creates environment variables for the child process
func (fm *FDManager) GenerateEnvVars() []string {
	var envVars []string

	for _, mapping := range fm.mappings {
		envVar := fmt.Sprintf("WW_FD_%s=%d", strings.ToUpper(mapping.name), mapping.targetFD)
		envVars = append(envVars, envVar)
	}

	return envVars
}

// Close cleans up all managed file descriptors
func (fm *FDManager) Close() error {
	var errors []string

	for _, mapping := range fm.mappings {
		if err := syscall.Close(mapping.sourceFD); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close fd %d: %v", mapping.sourceFD, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing FDs: %s", strings.Join(errors, "; "))
	}

	return nil
}
