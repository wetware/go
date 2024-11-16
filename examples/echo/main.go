//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"flag"
	"io"
	"os"

	"github.com/wetware/go/std/system"
)

//export echo
func echo() {
	if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}

func main() {
	stdin := flag.Bool("stdin", false, "read from standard input")
	serve := flag.Bool("serve", false, "handle async method calls")
	flag.Parse()

	if *stdin {
		echo()
	}

	if *serve {
		// Yield control to the scheduler.
		os.Exit(system.StatusAwaiting)
		// The caller will intercept interface{ExitCode() uint32} and
		// check if e.ExitCode() == system.StatusAwaiting.
		//
		// The top-level command will block until the runtime context
		// expires.
	}

	// Implicit status code 0 works as expected.
	// Caller will resolve to err = nil.
	// Top-level CLI command will unblock.
}
