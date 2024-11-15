//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"flag"
	"io"
	"os"
)

//export echo
func echo() {
	if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}

func main() {
	stdin := flag.Bool("stdin", false, "read from standard input")
	serve := flag.Bool("serve", false, "handle async method calls")
	flag.Parse()

	if *stdin {
		echo()
	}

	if *serve {
		// Signal to caller that this module is ready to handle
		// incoming method calls.
		os.Exit(0x00ff0000)
	}
}
