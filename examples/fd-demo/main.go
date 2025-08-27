package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	fmt.Println("File Descriptor Demo")
	fmt.Println("====================")

	// Check for RPC socket (always available at FD 3)
	if os.NewFile(3, "rpc") != nil {
		fmt.Println("✓ RPC socket available at FD 3")
	} else {
		fmt.Println("✗ RPC socket not available at FD 3")
	}

	// Check for user-provided file descriptors
	envVars := os.Environ()
	for _, env := range envVars {
		if len(env) > 6 && env[:6] == "WW_FD_" {
			fmt.Printf("✓ %s\n", env)

			// Extract the FD number and try to use it
			parts := []rune(env)
			fdStr := ""
			for i := 6; i < len(parts); i++ {
				if parts[i] == '=' {
					fdStr = string(parts[i+1:])
					break
				}
			}

			if fdStr != "" {
				if fd, err := strconv.Atoi(fdStr); err == nil {
					if file := os.NewFile(uintptr(fd), "user-fd"); file != nil {
						// Try to get file info
						if stat, err := file.Stat(); err == nil {
							fmt.Printf("  └─ File: %s, Size: %d bytes\n", stat.Name(), stat.Size())
						} else {
							fmt.Printf("  └─ File: %s (stat failed: %v)\n", file.Name(), err)
						}
						file.Close()
					} else {
						fmt.Printf("  └─ Failed to open FD %d\n", fd)
					}
				}
			}
		}
	}

	// If no user FDs provided, show a helpful message
	hasUserFDs := false
	for _, env := range envVars {
		if len(env) > 6 && env[:6] == "WW_FD_" {
			hasUserFDs = true
			break
		}
	}

	if !hasUserFDs {
		fmt.Println("\nNo user file descriptors provided.")
		fmt.Println("Try running with: ww run --with-fd demo=3 --with-fd log=4 ./fd-demo")
	}

	fmt.Println("\nDemo completed.")
}
