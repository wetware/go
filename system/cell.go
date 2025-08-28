package system

import (
	"context"
	"os"
	"syscall"

	capnp "capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
)

const BOOTSTRAP_FD = 3

// SocketConfig configures a pair of Unix domain sockets for IPC between
// a host process and guest process.
type SocketConfig[T ~capnp.ClientKind] struct {
	// BootstrapClient is the client that will be passed down to the guest
	// process. It is used to bootstrap the guest process's RPC connection.
	//
	// The socket steals the reference to T, and releases it when the conn
	// is closed.
	BootstrapClient T
}

// New creates a pair of connected Unix domain sockets. The host socket is
// returned first, followed by the guest socket. The caller is responsible
// for closing both sockets.
func (c SocketConfig[T]) New(ctx context.Context) (*rpc.Conn, *os.File, error) {
	host, guest, err := c.Socketpair()
	if err != nil {
		return nil, nil, err
	}

	conn := rpc.NewConn(rpc.NewStreamTransport(host), c.NewRPCOptions(ctx))
	return conn, guest, nil
}

func (c SocketConfig[T]) NewRPCOptions(ctx context.Context) *rpc.Options {
	return &rpc.Options{
		BaseContext:     func() context.Context { return ctx },
		BootstrapClient: capnp.Client(c.BootstrapClient),
	}
}

// socketpair creates a pair of connected Unix domain sockets for bidirectional communication
func (SocketConfig[T]) Socketpair() (*os.File, *os.File, error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, nil, err
	}

	host := os.NewFile(uintptr(fds[0]), "host")
	guest := os.NewFile(uintptr(fds[1]), "guest")

	return host, guest, nil
}
