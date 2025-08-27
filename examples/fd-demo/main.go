package main

import (
	"fmt"
	"os"

	"github.com/wetware/go/util"
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

	// Use utility function to get FD information
	fdMap := util.GetFDMap()
	if len(fdMap) > 0 {
		fmt.Println("\nFile Descriptors Available:")
		for name, fd := range fdMap {
			fmt.Printf("✓ %s -> FD %d\n", name, fd)

			// Try to get file info directly from the FD
			if file := os.NewFile(uintptr(fd), name); file != nil {
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
	} else {
		fmt.Println("\nNo user file descriptors provided.")
		fmt.Println("Try running with: ww run --with-fd demo=3 --with-fd log=4 ./fd-demo")
	}

	fmt.Println("\nDemo completed.")
}
