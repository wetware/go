//go:build wasm
// +build wasm

//go:generate tinygo build -o main.wasm -target wasm .

package main

import (
	"fmt"
	"os"
)

// This is a simple echo program that can be compiled to WASM
func main() {
	// Read from stdin and echo to stdout
	buf := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}
		if n > 0 {
			os.Stdout.Write(buf[:n])
		}
	}
}

// Echo function that can be called from the WASM module
//
//export echo
func echo(input string) string {
	return fmt.Sprintf("Echo: %s", input)
}
