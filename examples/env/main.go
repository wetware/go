//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"fmt"
	"os"
)

func main() {
	for _, v := range os.Environ() {
		fmt.Fprintln(os.Stdout, v)
	}
}
