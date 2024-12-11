//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"

	"github.com/wetware/go/proc"
)

func main() {}

//export deliver
func deliver(hdrLen uint32) {
	var r = io.LimitedReader{
		R: os.Stdin,
		N: int64(hdrLen),
	}

	var c proc.Call
	if err := json.NewDecoder(&r).Decode(&c); err != nil {
		slog.Error("failed to read header",
			"reason", err)
		return
	}

	// f, err := os.OpenFile(name, os.O_WRONLY, 0)
	// if err != nil {
	// 	slog.Error("failed to open file",
	// 		"path", name,
	// 		"reason", err)
	// 	return
	// }
	// defer f.Close()

	// r.N = int64(bodyLen)
	// io.Copy(f, r)
}
