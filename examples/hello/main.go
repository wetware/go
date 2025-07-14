//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"io"
	"os"
	"strings"
)

func main() {
	const message = "Hello, Wetware!\n"

	status := 0
	if n, err := io.Copy(os.Stdout, strings.NewReader(message)); err != nil {
		status = int(n)
		report := err.Error()
		io.Copy(os.Stderr, strings.NewReader(report))
	}

	os.Exit(status)
}
