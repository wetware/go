//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"io"
	"os"
)

// main is the entry point for synchronous mode.
// It processes one complete message from stdin and exits.
func main() {
	// Echo: copy stdin to stdout using io.Copy
	// io.Copy uses an internal 32KB buffer by default
	if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
		os.Stderr.WriteString("Error copying stdin to stdout: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer os.Stdout.Sync()
	// implicitly returns 0 to indicate successful completion
}

// poll is the async entry point for stream-based processing.
// This function is called by the wetware runtime when a new stream
// is established for this process.
//
//export poll
func poll() {
	// In async mode, we process each incoming stream
	// This is the same logic as main() but for individual streams
	if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
		os.Stderr.WriteString("Error in poll: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer os.Stdout.Sync()
}
