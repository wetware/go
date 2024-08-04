//go:generate env GOOS=wasip1 GOARCH=wasm go build -o main.wasm

package main

import (
	"fmt"
	"os"
)

func main() {
	b, err := os.ReadFile("testdata")
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Print test data to stdout.
	// If the contents of ./testdata grows to exceed
	// 16 bytes, make sure to expand the buffer.
	buf := make([]byte, 16)
	buf = buf[:copy(buf, b)]
	fmt.Fprint(os.Stdout, string(buf))
}
