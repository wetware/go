//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(os.Args[0])
}
