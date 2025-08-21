package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// Check if we're running as a child process with fd environment variables
	var hasWWFDs bool
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "WW_FD_") {
			hasWWFDs = true
			break
		}
	}

	if hasWWFDs {
		fmt.Println("=== Child Process with FD Passing ===")

		// Show individual fd variables
		fmt.Println("Individual FD variables:")
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "WW_FD_") {
				fmt.Printf("  %s\n", env)
			}
		}

		// Try to access the passed file descriptors
		fmt.Println("\nAttempting to access passed file descriptors:")

		// Check for common fd names
		fdNames := []string{"DB", "CACHE", "LOGS", "INPUT", "OUTPUT"}
		for _, name := range fdNames {
			if fdVar := os.Getenv("WW_FD_" + name); fdVar != "" {
				fmt.Printf("  %s: fd %s\n", name, fdVar)

				// Try to open the fd to verify access (on macOS use /dev/fd/)
				if fd, err := os.OpenFile("/dev/fd/"+fdVar, os.O_RDONLY, 0); err == nil {
					fd.Close()
					fmt.Printf("    ✓ Successfully accessed fd %s\n", fdVar)
				} else {
					fmt.Printf("    ✗ Failed to access fd %s: %v\n", fdVar, err)
				}
			}
		}

		// Show current working directory and jail info
		if cwd, err := os.Getwd(); err == nil {
			fmt.Printf("\nCurrent working directory: %s\n", cwd)
		}

		// List files in current directory (jail)
		if entries, err := os.ReadDir("."); err == nil {
			fmt.Println("\nFiles in jail directory:")
			for _, entry := range entries {
				info, _ := entry.Info()
				if info != nil {
					fmt.Printf("  %s (%s)\n", entry.Name(), info.Mode().String())
				} else {
					fmt.Printf("  %s\n", entry.Name())
				}
			}
		}

		return
	}

	// Parent process - show usage information
	fmt.Println("=== Simple FD Passing Demo ===")
	fmt.Println("This program demonstrates the simplified file descriptor passing capabilities.")
	fmt.Println()
	fmt.Println("To test fd passing, run this program with ww run:")
	fmt.Println()
	fmt.Println("Example 1: Single file descriptor")
	fmt.Println("  echo 'Hello World' > test.txt")
	fmt.Println("  ww run --fd input=3 3<test.txt /tmp/fd-demo")
	fmt.Println()
	fmt.Println("Example 2: Multiple file descriptors")
	fmt.Println("  echo 'data' > input.txt")
	fmt.Println("  touch output.txt")
	fmt.Println("  ww run --fd input=3 --fd output=4 3<input.txt 4>output.txt /tmp/fd-demo")
	fmt.Println()
	fmt.Println("Example 3: Database and cache fds")
	fmt.Println("  ww run --fd db=3 --fd cache=4 3<db.txt 4<cache.txt /tmp/fd-demo")
	fmt.Println()
	fmt.Println("The child process will receive environment variables:")
	fmt.Println("  WW_FD_INPUT=10   (auto-assigned)")
	fmt.Println("  WW_FD_OUTPUT=11  (auto-assigned)")
	fmt.Println("  WW_FD_DB=12      (auto-assigned)")
	fmt.Println("  WW_FD_CACHE=13   (auto-assigned)")
	fmt.Println()
	fmt.Println("File descriptor numbers are automatically assigned starting at 10.")
}
