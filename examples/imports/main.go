package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// Check if we're running as a child process with fd environment variables
	if wwFDS := os.Getenv("WW_FDS"); wwFDS != "" {
		fmt.Println("=== Child Process with FD Passing ===")
		fmt.Printf("WW_FDS: %s\n", wwFDS)

		// Show individual fd variables
		fmt.Println("\nIndividual FD variables:")
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

				// Try to open the fd to verify access
				if fd, err := os.OpenFile("/proc/self/fd/"+fdVar, os.O_RDONLY, 0); err == nil {
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
	fmt.Println("=== FD Passing Demo ===")
	fmt.Println("This program demonstrates file descriptor passing capabilities.")
	fmt.Println()
	fmt.Println("To test fd passing, run this program with ww run:")
	fmt.Println()
	fmt.Println("Example 1: Lisp capability imports (recommended)")
	fmt.Println("  cat >imports <<'SEXP'")
	fmt.Println("  (")
	fmt.Println("    (fd :name \"input\" :fd 3 :mode \"r\" :target 10)")
	fmt.Println("    (fd :name \"output\" :fd 4 :mode \"w\" :target 11)")
	fmt.Println("  )")
	fmt.Println("  SEXP")
	fmt.Println("  ww run --fd-from @imports 3<test.txt 4>output.txt /tmp/fd-demo")
	fmt.Println()
	fmt.Println("Example 2: Command-line arguments")
	fmt.Println("  echo 'Hello World' > test.txt")
	fmt.Println("  ww run --fd input=3,mode=r,type=file 3<test.txt /tmp/fd-demo")
	fmt.Println()
	fmt.Println("Example 3: Multiple fds")
	fmt.Println("  mkfifo db.sock")
	fmt.Println("  touch app.log")
	fmt.Println("  ww run \\")
	fmt.Println("    --fd db=3,mode=rw,type=socket,target=10 \\")
	fmt.Println("    --fd logs=4,mode=w,type=file,target=11 \\")
	fmt.Println("    3<>db.sock 4>app.log \\")
	fmt.Println("    /tmp/fd-demo")
	fmt.Println()
	fmt.Println("Example 4: With verbose logging")
	fmt.Println("  ww run --fd-verbose --fd input=3,mode=r,type=file 3<test.txt /tmp/fd-demo")
	fmt.Println()
	fmt.Println("Example 5: Inherit existing fd")
	fmt.Println("  exec 3<test.txt")
	fmt.Println("  ww run --fdctl inherit:3 /tmp/fd-demo")
	fmt.Println()
	fmt.Println("The program will show different output when run with fd passing.")
	fmt.Println("Check the environment variables and file descriptor access.")
}
