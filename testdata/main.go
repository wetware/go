//go:generate env GOOS=wasip1 GOARCH=wasm go build -o main.wasm

package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	n, err := io.Copy(os.Stdout, os.Stdin)
	fmt.Fprintf(os.Stderr, "copied %d bytes", n)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
