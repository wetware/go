//go:generate mockgen -source=glia.go -destination=mock_test.go -package=glia_test

package glia

import (
	"context"
	"io"
	"log/slog"

	"github.com/tetratelabs/wazero/api"
	"github.com/wetware/go/proc"
)

type Env interface {
	Log() *slog.Logger
}

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
	ProcID() string
	MethodName() string
	io.ReadWriteCloser
	CloseRead() error
	CloseWrite() error
}
