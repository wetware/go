//go:generate mockgen -source=glia.go -destination=mock_test.go -package=glia_test

package glia

import (
	"io"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	core_routing "github.com/libp2p/go-libp2p/core/routing"
)

type Env interface {
	Log() *slog.Logger
	LocalHost() host.Host
	Routing() core_routing.Routing
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
