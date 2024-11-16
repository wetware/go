//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/wetware/go/std/system"
)

func main() {
	if nargs := len(os.Args); nargs < 1 {
		slog.Error("wrong number of arguments",
			"want", 1,
			"got", nargs,
			"args", os.Args)
		os.Exit(system.StatusInvalidArgs)
	}

	f, err := os.Open(os.Args[0])
	if err != nil {
		slog.Error("failed to open file",
			"reason", err,
			"name", os.Args[0])
		os.Exit(system.StatusInvalidArgs)
	}
	defer f.Close()

	var n int64
	var status int
	if n, err = io.Copy(f, os.Stdin); err != nil {
		status = system.StatusFailed
		err = fmt.Errorf("request: %w", err)
	} else if n, err = io.Copy(os.Stdout, f); err != nil {
		status = system.StatusFailed
		err = fmt.Errorf("response: %w", err)
	}

	if err != nil {
		slog.Error("failed to read message from stdin",
			"reason", err,
			"read", n)
	}

	os.Exit(status)
}
