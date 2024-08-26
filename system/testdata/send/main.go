//go:generate env GOOS=wasip1 GOARCH=wasm go build -o main.wasm

package main

import (
	"log"
	"os"
)

func main() {
	const path = "/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N/ww/8D9E923E-9F17-4869-8DED-6725D22C4F7F"

	// Open the file for writing.  This creates a new buffer
	// for the message.  The message will be sent when f.Close()
	// is called.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close() // sends the data to the remote process

	// Write some text data to the buffer
	_, err = f.WriteString("Hello, world!\nThis is some text data.")
	if err != nil {
		log.Fatal(err)
	}
}
