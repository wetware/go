package run

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// FDConfig represents a simple file descriptor configuration
type FDConfig struct {
	Name     string // Name of the fd (e.g., "db", "cache")
	SourceFD int    // Source file descriptor number
	TargetFD int    // Target file descriptor number in child
}

// FDManager handles basic file descriptor operations
type FDManager struct {
	configs map[string]*FDConfig
	files   []*os.File
}

// NewFDManager creates a new FD manager
func NewFDManager() *FDManager {
	return &FDManager{
		configs: make(map[string]*FDConfig),
		files:   nil,
	}
}

// ParseFDFlag parses the --with-fd flag value in simple "name=fdnum" format
func (fm *FDManager) ParseFDFlag(value string) error {
	parts := strings.Split(value, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid --with-fd format: must be name=fdnum")
	}

	name := parts[0]
	fdnum, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid fd number: %s", parts[1])
	}

	// Check for duplicate names
	if _, exists := fm.configs[name]; exists {
		return fmt.Errorf("duplicate fd name: %s", name)
	}

	config := &FDConfig{
		Name:     name,
		SourceFD: fdnum,
		TargetFD: 0, // Will be auto-assigned
	}

	fm.configs[name] = config
	return nil
}

// PrepareFDs prepares file descriptors for the child process
func (fm *FDManager) PrepareFDs() ([]*os.File, error) {
	var files []*os.File

	// Process each fd configuration first to create the files
	for _, config := range fm.configs {
		file, err := fm.prepareFD(config)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare fd %s: %w", config.Name, err)
		}

		if file != nil {
			files = append(files, file)
		}
	}

	// Now assign target FDs based on the actual order in ExtraFiles
	// The child will receive these starting at FD 3 (after stdin=0, stdout=1, stderr=2)
	nextFD := 3
	counter := 0
	for _, config := range fm.configs {
		config.TargetFD = nextFD + counter
		counter++
	}

	fm.files = files
	return files, nil
}

// prepareFD prepares a single file descriptor
func (fm *FDManager) prepareFD(config *FDConfig) (*os.File, error) {
	sourceFile := os.NewFile(uintptr(config.SourceFD), config.Name)
	if sourceFile == nil {
		return nil, fmt.Errorf("invalid source fd: %d", config.SourceFD)
	}

	// Duplicate the file descriptor
	dupFD, err := fm.dupFD(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to duplicate fd: %w", err)
	}

	return dupFD, nil
}

// dupFD duplicates a file descriptor
func (fm *FDManager) dupFD(file *os.File) (*os.File, error) {
	fd := int(file.Fd())
	newFD, err := syscall.Dup(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to dup fd %d: %w", fd, err)
	}

	newFile := os.NewFile(uintptr(newFD), file.Name())
	if newFile == nil {
		syscall.Close(newFD)
		return nil, fmt.Errorf("failed to create file from dup'd fd %d", newFD)
	}

	return newFile, nil
}

// GenerateEnvVars generates environment variables for the child process
func (fm *FDManager) GenerateEnvVars() []string {
	var envVars []string

	// Generate individual WW_FD_<NAME> variables
	for _, config := range fm.configs {
		envVar := fmt.Sprintf("WW_FD_%s=%d", strings.ToUpper(config.Name), config.TargetFD)
		envVars = append(envVars, envVar)
	}

	return envVars
}

// Close closes all managed file descriptors
func (fm *FDManager) Close() error {
	var errors []error

	for _, file := range fm.files {
		if err := file.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing fds: %v", errors)
	}

	return nil
}
