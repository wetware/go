//go:generate mockgen -source=glia.go -destination=mock_test.go -package=glia_test

package glia

import (
	"context"
	"io"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tetratelabs/wazero/api"
	"github.com/wetware/go/proc"
)

type Router interface {
	GetProc(pid string) (Proc, error)
}

type Proc interface {
	Reserve(context.Context, io.ReadWriteCloser) error
	Release()

	api.Closer
	String() string
	Method(name string) proc.Method
}

type Stream interface {
	Protocol() protocol.ID
	Destination() string
	ProcID() string
	MethodName() string

	io.ReadWriteCloser
	CloseRead() error
	CloseWrite() error
}
