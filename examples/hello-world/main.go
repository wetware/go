//go:generate GOOS=wasip1 GOARCH=wasm go build -o main.wasm

package main

import "fmt"

func main() {
	fmt.Println("Hello, Wetware!")
}
