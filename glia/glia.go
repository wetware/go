//go:generate mockgen -source=glia.go -destination=mock_test.go -package=glia_test

package glia

import (
	"context"
	"io"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	core_routing "github.com/libp2p/go-libp2p/core/routing"
	"github.com/tetratelabs/wazero/api"
	"github.com/wetware/go/proc"
)

type Env interface {
	Log() *slog.Logger
	LocalHost() host.Host
	Routing() core_routing.Routing
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
