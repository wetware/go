//go:generate env GOOS=wasip1 GOARCH=wasm go build -o main.wasm

package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("nop")
		time.Sleep(time.Second)
	}
}
