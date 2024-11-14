//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"fmt"
	"io"
	"os"
)

//export environment
func environment(size uint32) {
	// Defensive:  clear out any data that the user may have sent,
	// even if we aren't expecting it.
	defer io.Copy(io.Discard, os.Stdin)

	// Calmly print the environment to stdout
	for _, v := range os.Environ() {
		fmt.Fprintln(os.Stdout, v)
	}
}

func main() {}
