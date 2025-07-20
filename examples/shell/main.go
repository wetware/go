//go:generate tinygo build -o main.wasm -target=wasi -scheduler=asyncify main.go

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("!")

	f, err := os.Open(".")
	if err != nil {
		panic(err)
	}
	defer f.Close()
}
