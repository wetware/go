//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

func main() {
	if nargs := len(os.Args); nargs < 1 {
		slog.Error("wrong number of arguments",
			"want", 1,
			"got", nargs,
			"args", os.Args)
		os.Exit(1)
	}

	f, err := os.Open(os.Args[0])
	if err != nil {
		slog.Error("failed to open file",
			"reason", err,
			"name", os.Args[0])
		os.Exit(1)
	}
	defer f.Close()

	var n int64
	if n, err = io.Copy(f, os.Stdin); err != nil {
		err = fmt.Errorf("request: %w", err)
	} else if n, err = io.Copy(os.Stdout, f); err != nil {
		err = fmt.Errorf("response: %w", err)
	}

	if err != nil {
		slog.Error("failed to read message from stdin",
			"reason", err,
			"read", n)
		os.Exit(2)
	}

	slog.Debug("delivered message",
		"size", n)
}
