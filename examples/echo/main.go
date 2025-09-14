//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"io"
	"log/slog"
	"os"
)

func main() {}

// Echo function that can be called from the WASM module
//
//export poll
func poll() {
	var buf [512]byte
	if n, err := io.CopyBuffer(os.Stdout, os.Stdin, buf[:]); err != nil {
		slog.Error("failed to copy", "reason", err, "written", n)
	}
}
