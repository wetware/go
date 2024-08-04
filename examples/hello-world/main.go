//go:generate env GOOS=wasip1 GOARCH=wasm go build -o main.wasm

package main

import (
	"os"
)

func main() {
	os.Exit(0)
}
