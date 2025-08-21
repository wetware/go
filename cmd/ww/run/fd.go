package run

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
)

// FDConfig represents a file descriptor configuration
type FDConfig struct {
	Name     string // Name of the fd (e.g., "db", "cache")
	SourceFD int    // Source file descriptor number
	TargetFD int    // Target file descriptor number in child
	Mode     string // Access mode: "r", "w", "rw"
	Type     string // Type: "stream", "file", "socket"
	Target   string // Target path for symlink creation
	Move     bool   // Whether to move (close original) or dup
	Cloexec  bool   // Whether to set CLOEXEC flag
	PathLink bool   // Whether to create symlink in jail
}

// FDManager handles file descriptor operations
type FDManager struct {
	configs map[string]*FDConfig
	files   []*os.File
	verbose bool
	env     core.Env // Slurp environment for Lisp validation
}

// NewFDManager creates a new FD manager
func NewFDManager(verbose bool) *FDManager {
	return &FDManager{
		configs: make(map[string]*FDConfig),
		files:   make([]*os.File, 0),
		verbose: verbose,
		env:     core.New(nil),
	}
}

// ParseFDFlag parses the --fd flag value
func (fm *FDManager) ParseFDFlag(value string) error {
	parts := strings.Split(value, ",")
	if len(parts) < 1 {
		return fmt.Errorf("invalid --fd format: empty value")
	}

	// Parse name=fdnum
	nameFD := strings.Split(parts[0], "=")
	if len(nameFD) != 2 {
		return fmt.Errorf("invalid --fd format: first part must be name=fdnum")
	}

	name := nameFD[0]
	fdnum, err := strconv.Atoi(nameFD[1])
	if err != nil {
		return fmt.Errorf("invalid fd number: %s", nameFD[1])
	}

	config := &FDConfig{
		Name:     name,
		SourceFD: fdnum,
		Mode:     "rw",   // default
		Type:     "file", // default
		Move:     false,  // default
		Cloexec:  true,   // default
		PathLink: false,  // default
	}

	// Parse additional options
	for i := 1; i < len(parts); i++ {
		opt := strings.Split(parts[i], "=")
		if len(opt) != 2 {
			return fmt.Errorf("invalid option format: %s", parts[i])
		}

		key, value := opt[0], opt[1]
		switch key {
		case "mode":
			config.Mode = value
		case "type":
			config.Type = value
		case "target":
			if targetFD, err := strconv.Atoi(value); err == nil {
				config.TargetFD = targetFD
			} else {
				config.Target = value
			}
		case "move":
			config.Move = value == "true"
		case "cloexec":
			config.Cloexec = value == "true"
		case "pathlink":
			config.PathLink = value == "true"
		default:
			return fmt.Errorf("unknown option: %s", key)
		}
	}

	// Validate mode
	if config.Mode != "r" && config.Mode != "w" && config.Mode != "rw" {
		return fmt.Errorf("invalid mode: %s (must be r, w, or rw)", config.Mode)
	}

	// Validate type
	if config.Type != "stream" && config.Type != "file" && config.Type != "socket" {
		return fmt.Errorf("invalid type: %s (must be stream, file, or socket)", config.Type)
	}

	// Check for duplicate names
	if _, exists := fm.configs[name]; exists {
		return fmt.Errorf("duplicate fd name: %s", name)
	}

	fm.configs[name] = config
	return nil
}

// ParseFDMapFlag parses the --fd-map flag value
func (fm *FDManager) ParseFDMapFlag(value string) error {
	parts := strings.Split(value, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid --fd-map format: must be name=targetfd")
	}

	name := parts[0]
	targetFD, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid target fd number: %s", parts[1])
	}

	config, exists := fm.configs[name]
	if !exists {
		return fmt.Errorf("fd name not found: %s", name)
	}

	config.TargetFD = targetFD
	return nil
}

// ParseFDCTLFlag parses the --fdctl flag value
func (fm *FDManager) ParseFDCTLFlag(value string) error {
	if strings.HasPrefix(value, "inherit:") {
		fdnumStr := strings.TrimPrefix(value, "inherit:")
		fdnum, err := strconv.Atoi(fdnumStr)
		if err != nil {
			return fmt.Errorf("invalid inherit fd number: %s", fdnumStr)
		}

		// Create a config for inherited fd
		config := &FDConfig{
			Name:     fmt.Sprintf("inherit_%d", fdnum),
			SourceFD: fdnum,
			Mode:     "rw",
			Type:     "file",
			Move:     false,
			Cloexec:  true,
			PathLink: false,
		}

		fm.configs[config.Name] = config
		return nil
	}

	// Handle unix socket path (SCM_RIGHTS)
	// This is a placeholder for future implementation
	return fmt.Errorf("unix socket SCM_RIGHTS not yet implemented")
}

// UseSystemdFDs imports file descriptors from systemd socket activation
func (fm *FDManager) UseSystemdFDs(prefix string) error {
	listenFDS := os.Getenv("LISTEN_FDS")
	listenPID := os.Getenv("LISTEN_PID")

	if listenFDS == "" || listenPID == "" {
		return fmt.Errorf("systemd socket activation not available")
	}

	// Verify we're the expected process
	ourPID := os.Getpid()
	expectedPID, err := strconv.Atoi(listenPID)
	if err != nil {
		return fmt.Errorf("invalid LISTEN_PID: %s", listenPID)
	}

	if ourPID != expectedPID {
		return fmt.Errorf("LISTEN_PID mismatch: expected %d, got %d", expectedPID, ourPID)
	}

	fdCount, err := strconv.Atoi(listenFDS)
	if err != nil {
		return fmt.Errorf("invalid LISTEN_FDS: %s", listenFDS)
	}

	// Import systemd-provided fds
	for i := 0; i < fdCount; i++ {
		fd := 3 + i // systemd starts at fd 3
		name := fmt.Sprintf("%s%d", prefix, i)

		config := &FDConfig{
			Name:     name,
			SourceFD: fd,
			Mode:     "rw",
			Type:     "socket",
			Move:     false,
			Cloexec:  true,
			PathLink: false,
		}

		fm.configs[name] = config
	}

	return nil
}

// ParseFDFromFile parses fd specifications from a file or stdin
func (fm *FDManager) ParseFDFromFile(path string) error {
	var data []byte
	var err error

	if path == "-" {
		data, err = io.ReadAll(os.Stdin) // stdin
	} else {
		path = strings.TrimPrefix(path, "@")
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return fmt.Errorf("failed to read fd spec: %w", err)
	}

	rd := reader.New(strings.NewReader(string(data)))
	expr, err := rd.One()
	if err != nil {
		return fmt.Errorf("failed to parse S-expression: %w", err)
	}

	// Process the parsed capability table
	return fm.processSExpr(expr)
}

// processSExpr processes a parsed capability table and extracts fd configurations
func (fm *FDManager) processSExpr(expr core.Any) error {
	// Check if it's a sequence (vector of fd specs)
	if seq, ok := expr.(core.Seq); ok {
		return fm.processFDVector(seq)
	}

	// Single fd spec
	return fm.processSingleFD(expr)
}

// processFDVector processes a vector of fd capabilities
func (fm *FDManager) processFDVector(seq core.Seq) error {
	for {
		// Get the first item
		item, err := seq.First()
		if err != nil {
			return fmt.Errorf("failed to get first item: %w", err)
		}

		// Process the item
		if err := fm.processSingleFD(item); err != nil {
			return fmt.Errorf("failed to process fd spec: %w", err)
		}

		// Get the next sequence
		next, err := seq.Next()
		if err != nil {
			return fmt.Errorf("failed to get next item: %w", err)
		}

		if next == nil {
			break // End of sequence
		}

		seq = next
	}

	return nil
}

// processSingleFD processes a single fd capability
func (fm *FDManager) processSingleFD(expr core.Any) error {
	// Check if it's a sequence (fd capability like (fd :name "stdin" :fd 0 :mode "r"))
	if seq, ok := expr.(core.Seq); ok {
		return fm.parseFDSpec(seq)
	}

	return nil // Skip non-sequence items
}

// parseFDSpec parses a single fd capability from a sequence
func (fm *FDManager) parseFDSpec(seq core.Seq) error {
	config := &FDConfig{
		Mode:     "rw",
		Type:     "file",
		Move:     false,
		Cloexec:  true,
		PathLink: false,
	}

	// Process the sequence to extract fd configuration
	// New format: {:name "stdin" :fd 0 :mode "r" :target 0}
	for {
		item, err := seq.First()
		if err != nil {
			break
		}

		// Check if it's a keyword (starts with :)
		if keyword, ok := item.(builtin.Keyword); ok {
			key := string(keyword)
			key = strings.TrimPrefix(key, ":")

			// Get the value (next item)
			next, err := seq.Next()
			if err != nil {
				return fmt.Errorf("failed to get value for key %s: %w", key, err)
			}

			if next == nil {
				return fmt.Errorf("missing value for key %s", key)
			}

			valueItem, err := next.First()
			if err != nil {
				return fmt.Errorf("failed to get value for key %s: %w", key, err)
			}

			// Apply the key-value to the config
			if err := fm.applyConfigValue(config, key, valueItem); err != nil {
				return fmt.Errorf("failed to apply %s: %w", key, err)
			}

			// Move to the sequence after the value
			seq = next
		}

		// Get next item
		next, err := seq.Next()
		if err != nil || next == nil {
			break
		}
		seq = next
	}

	// Validate the configuration
	if config.Name == "" {
		return fmt.Errorf("missing name in fd spec")
	}

	// Check for duplicate names
	if _, exists := fm.configs[config.Name]; exists {
		return fmt.Errorf("duplicate fd name: %s", config.Name)
	}

	fm.configs[config.Name] = config
	return nil
}

// applyConfigValue applies a key-value pair to the config
func (fm *FDManager) applyConfigValue(config *FDConfig, key string, value core.Any) error {
	// Convert value to string
	valueStr := fmt.Sprintf("%v", value)

	// Remove quotes from value if present
	valueStr = strings.Trim(valueStr, "\"")

	// Apply the key-value to the config
	switch key {
	case "name":
		config.Name = valueStr
	case "fd":
		if fd, err := strconv.Atoi(valueStr); err == nil {
			config.SourceFD = fd
		}
	case "mode":
		config.Mode = valueStr
	case "type":
		config.Type = valueStr
	case "target":
		if targetFD, err := strconv.Atoi(valueStr); err == nil {
			config.TargetFD = targetFD
		} else {
			config.Target = valueStr
		}
	}

	return nil
}

// parseKeyValue parses a key-value pair from a sequence
func (fm *FDManager) parseKeyValue(seq core.Seq, config *FDConfig) error {
	// Get the key (first item)
	keyItem, err := seq.First()
	if err != nil {
		return err
	}

	// Get the value (second item)
	next, err := seq.Next()
	if err != nil {
		return err
	}

	if next == nil {
		return fmt.Errorf("missing value for key")
	}

	valueItem, err := next.First()
	if err != nil {
		return err
	}

	// Convert to strings
	key := fmt.Sprintf("%v", keyItem)
	value := fmt.Sprintf("%v", valueItem)

	// Handle keyword-style keys (remove leading colon)
	key = strings.TrimPrefix(key, ":")

	// Remove quotes from value if present
	value = strings.Trim(value, "\"")

	// Apply the key-value to the config
	switch key {
	case "name":
		config.Name = value
	case "fd":
		if fd, err := strconv.Atoi(value); err == nil {
			config.SourceFD = fd
		}
	case "mode":
		config.Mode = value
	case "type":
		config.Type = value
	case "target":
		if targetFD, err := strconv.Atoi(value); err == nil {
			config.TargetFD = targetFD
		} else {
			config.Target = value
		}
	}

	return nil
}

// PrepareFDs prepares file descriptors for the child process
func (fm *FDManager) PrepareFDs() ([]*os.File, error) {
	var files []*os.File
	var targetFDs []int

	// Auto-assign target fds if not specified
	nextFD := 10 // Start at 10 to avoid conflicts with stdio
	for _, config := range fm.configs {
		if config.TargetFD == 0 {
			config.TargetFD = nextFD
			nextFD++
		}
		targetFDs = append(targetFDs, config.TargetFD)
	}

	// Check for duplicate target fds
	seen := make(map[int]bool)
	for _, targetFD := range targetFDs {
		if seen[targetFD] {
			return nil, fmt.Errorf("duplicate target fd: %d", targetFD)
		}
		seen[targetFD] = true
	}

	// Process each fd configuration
	for _, config := range fm.configs {
		file, err := fm.prepareFD(config)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare fd %s: %w", config.Name, err)
		}

		if file != nil {
			files = append(files, file)
			if fm.verbose {
				fmt.Fprintf(os.Stderr, "grant fd name=%s num=%d type=%s mode=%s move=%t cloexec=%t\n",
					config.Name, config.TargetFD, config.Type, config.Mode, config.Move, config.Cloexec)
			}
		}
	}

	fm.files = files
	return files, nil
}

// prepareFD prepares a single file descriptor
func (fm *FDManager) prepareFD(config *FDConfig) (*os.File, error) {
	// Get the source file
	sourceFile := os.NewFile(uintptr(config.SourceFD), config.Name)
	if sourceFile == nil {
		return nil, fmt.Errorf("invalid source fd: %d", config.SourceFD)
	}

	// Validate access mode
	if err := fm.validateAccess(sourceFile, config.Mode); err != nil {
		return nil, fmt.Errorf("access validation failed: %w", err)
	}

	// Infer type if not specified
	if config.Type == "" {
		config.Type = fm.inferType(sourceFile)
	}

	// Set CLOEXEC if requested
	if config.Cloexec {
		if err := fm.setCloexec(sourceFile); err != nil {
			return nil, fmt.Errorf("failed to set CLOEXEC: %w", err)
		}
	}

	return sourceFile, nil
}

// validateAccess validates that the file descriptor supports the requested access mode
func (fm *FDManager) validateAccess(file *os.File, mode string) error {
	// Try to invoke a Lisp validation function if available
	if validator, err := fm.env.Resolve("validate-fd-access"); err == nil {
		// Call the Lisp function with file info and mode
		fileInfo, _ := file.Stat()
		result, err := core.Eval(fm.env, nil, []core.Any{validator, fileInfo, mode})
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Check if validation passed (assuming result is boolean or nil)
		if result != nil && result != false {
			return nil
		}
		return fmt.Errorf("validation rejected fd access")
	}

	// Fallback to basic validation if no Lisp function available
	return nil
}

// SetValidator registers a Lisp function for file descriptor access validation
func (fm *FDManager) SetValidator(name string, validator core.Any) {
	fm.env.Bind(name, validator)
}

// inferType infers the type of a file descriptor
func (fm *FDManager) inferType(file *os.File) string {
	// Try to get file info to determine type
	info, err := file.Stat()
	if err != nil {
		return "file" // default fallback
	}

	mode := info.Mode()
	if mode.IsDir() {
		return "file"
	}

	// Check if it's a socket
	if mode&os.ModeSocket != 0 {
		return "socket"
	}

	// Check if it's a named pipe
	if mode&os.ModeNamedPipe != 0 {
		return "stream"
	}

	return "file"
}

// setCloexec sets the CLOEXEC flag on a file descriptor
func (fm *FDManager) setCloexec(file *os.File) error {
	// For now, we'll skip setting CLOEXEC as it requires platform-specific syscalls
	// This can be enhanced later with proper platform detection
	return nil
}

// CreateSymlinks creates symlinks in the jail directory if requested
func (fm *FDManager) CreateSymlinks(jailDir string) error {
	for _, config := range fm.configs {
		if config.PathLink && config.Target != "" {
			linkPath := filepath.Join(jailDir, config.Target)
			targetPath := fmt.Sprintf("/proc/self/fd/%d", config.TargetFD)

			if err := os.Symlink(targetPath, linkPath); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %w", linkPath, targetPath, err)
			}
		}
	}
	return nil
}

// GenerateEnvVars generates environment variables for the child process
func (fm *FDManager) GenerateEnvVars() []string {
	var envVars []string

	// Generate WW_FDS S-expression
	wwFDS := fm.generateWWFDS()
	envVars = append(envVars, fmt.Sprintf("WW_FDS=%s", wwFDS))

	// Generate individual WW_FD_<NAME> variables
	for _, config := range fm.configs {
		envVar := fmt.Sprintf("WW_FD_%s=%d", strings.ToUpper(config.Name), config.TargetFD)
		envVars = append(envVars, envVar)
	}

	return envVars
}

// generateWWFDS generates the WW_FDS S-expression
func (fm *FDManager) generateWWFDS() string {
	var parts []string

	for _, config := range fm.configs {
		fdSpec := fmt.Sprintf("(fd :name \"%s\" :fd %d :mode \"%s\" :type \"%s\"",
			config.Name, config.TargetFD, config.Mode, config.Type)

		if config.Target != "" {
			fdSpec += fmt.Sprintf(" :target \"%s\"", config.Target)
		}

		fdSpec += ")"
		parts = append(parts, fdSpec)
	}

	return fmt.Sprintf("(%s)", strings.Join(parts, " "))
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
