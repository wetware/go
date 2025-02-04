//go:generate mockgen -source=glia.go -destination=mock_test.go -package=glia_test
//go:generate capnp compile -I$GOPATH/src/capnproto.org/go/capnp/std -ogo glia.capnp

package glia

import (
	"bytes"
	"context"
	"io"
	"log/slog"

	"github.com/tetratelabs/wazero/api"
	"github.com/wetware/go/proc"
)

type Proc interface {
	Reserve(ctx context.Context, body io.Reader) error
	Release()

	OutBuffer() *bytes.Reader

	api.Closer
	String() string
	Method(name string) proc.Method
}

type Router interface {
	GetProc(pid string) (Proc, error)
}

type Env interface {
	Log() *slog.Logger
}
