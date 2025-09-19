package system

import (
	"context"
	"io"

	"github.com/libp2p/go-libp2p/core/protocol"
	"golang.org/x/sync/semaphore"
)

type Endpoint struct {
	Name string
	io.ReadWriteCloser
	sem *semaphore.Weighted
}

// Read implements io.Reader for Endpoint
func (e *Endpoint) Read(p []byte) (n int, err error) {
	if e.ReadWriteCloser == nil {
		// If no stream is available, return EOF immediately
		// This allows the WASM module to complete its main() function
		return 0, io.EOF
	}
	return e.ReadWriteCloser.Read(p)
}

// Write implements io.Writer for Endpoint
func (e *Endpoint) Write(p []byte) (n int, err error) {
	if e.ReadWriteCloser == nil {
		// If no stream is available, discard output
		return len(p), nil
	}
	return e.ReadWriteCloser.Write(p)
}

// String returns the full protocol identifier including the /ww/0.1.0/ prefix.
func (e Endpoint) String() string {
	proto := e.Protocol()
	return string(proto)
}

// Protocol returns the libp2p protocol ID for this endpoint.
func (e Endpoint) Protocol() protocol.ID {
	return protocol.ID("/ww/0.1.0/" + e.Name)
}

func (e *Endpoint) Close(context.Context) (err error) {
	if e.ReadWriteCloser != nil {
		err = e.ReadWriteCloser.Close()
	}
	return
}
