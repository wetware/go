//go:generate env GOOS=wasip1 GOARCH=wasm go build -o main.wasm

package main

import (
	"io"
	"os"
)

func main() {
	io.Copy(os.Stdout, os.Stdin)
}
