//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

func main() {
	if len(os.Args) < 1 {
		slog.Error("expected 1 argument, got 0",
			"status", 1)
		os.Exit(1)
	}

	f, err := os.Open(os.Args[0])
	if err != nil {
		slog.Error("failed to open file",
			"reason", err,
			"status", 2,
			"name", os.Args[0])
		os.Exit(2)
	}
	defer f.Close()

	var n int64
	var status int
	if n, err = io.Copy(f, os.Stdin); err != nil {
		status = 3
		err = fmt.Errorf("request: %w", err)
	} else if n, err = io.Copy(os.Stdout, f); err != nil {
		status = 4
		err = fmt.Errorf("response: %w", err)
	}

	if err != nil {
		slog.Error("failed to read message from stdin",
			"reason", err,
			"status", status,
			"read", n)
	}

	os.Exit(status)
}
