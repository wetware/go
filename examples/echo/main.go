//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"io"
	"os"
)

//export echo
func echo() {
	if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}

func main() {}
