//go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"unsafe"
)

var lim = io.LimitedReader{R: os.Stdin}

func main() {}

//go:wasm-module ww
//export deliver
func ww_deliver(size uint32) {
	lim.N = int64(size)

	b, err := io.ReadAll(&lim)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(os.Stdout, string(b))

	// Since this is a mock for a test, we just echo
	// the input back via stdout.
	ww_send(b)
}

func ww_send(b []byte) {
	ptr, size := bytesToPtr(b)
	_ww_send(ptr, size)
	runtime.KeepAlive(b)
}

//go:wasmimport ww send
func _ww_send(ptr, size uint32)

// bytesToPtr returns a pointer and size pair for the given data in a way
// compatible with WebAssembly numeric types.
// The returned pointer aliases the data hence the data must be kept alive
// until ptr is no longer needed.
func bytesToPtr(b []byte) (uint32, uint32) {
	ptr := unsafe.Pointer(unsafe.SliceData(b))
	return uint32(uintptr(ptr)), uint32(len(b))
}
