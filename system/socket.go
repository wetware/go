package system

import (
	context "context"
	"errors"
	"io"
	"math"

	"github.com/tetratelabs/wazero/api"
)

type SocketConfig struct {
	System interface{ Mailbox() io.Writer }
	Guest  interface{ ExportedFunction(string) api.Function }
}

func (b SocketConfig) Spawn() Proc {
	return Proc_ServerToClient(b.Build())
}

func (c SocketConfig) Build() (sock Socket) {
	sock.Buffer = c.System.Mailbox()
	sock.Deliver = c.Guest.ExportedFunction("deliver")
	return
}

type Socket struct {
	// Buffer is an io.Writer that accumulates bytes in preparation for
	// a call to deliver.  Default implementation is bytes.Buffer.
	Buffer io.Writer

	// Deliver is a guest export that takes a u32 number of bytes to
	// consume from Buffer.
	Deliver api.Function
}

func (p Socket) Handle(ctx context.Context, call Proc_handle) error {
	if p.Deliver == nil {
		return errors.New("missing export: deliver")
	}

	b, err := call.Args().Event()
	if err != nil {
		return err
	}

	n, err := p.Buffer.Write(b)
	if err != nil {
		return err
	} else if n > math.MaxUint32 {
		return errors.New("message size overflows u32")
	}

	return p.flush(ctx, uint32(n))
}

func (p Socket) flush(ctx context.Context, size uint32) error {
	_, err := p.Deliver.Call(ctx, api.EncodeU32(size))
	return err
}
