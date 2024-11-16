//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"bufio"
	"flag"
	"io"
	"log/slog"
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
	stdin := flag.Bool("stdin", false, "read data from stdin")
	flag.Parse()

	if *stdin {
		if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
			slog.Error("failed echo stdin",
				"reason", err)
			os.Exit(1)
		}

	} else {
		w := bufio.NewWriter(os.Stdout)
		for _, arg := range os.Args[1:] {
			w.WriteString(arg)
			w.WriteString(" ")
			if err := w.Flush(); err != nil {
				slog.Error("failed to flush argument to stdout",
					"reason", err)
				os.Exit(1)
			}
		}
	}

	if serve() {
		// Yield control to the scheduler.
		os.Exit(system.StatusAsync)
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

func serve() bool {
	switch os.Getenv("WW_SERVE") {
	case "", "false", "0":
		return false
	default:
		return true
	}
}
