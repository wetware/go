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
type SocketConfig struct {
	Membrane Importer_Server
}

// New creates a pair of connected Unix domain sockets. The host socket is
// returned first, followed by the guest socket. The caller is responsible
// for closing both sockets.
func (c SocketConfig) New(ctx context.Context) (*rpc.Conn, *os.File, error) {
	host, guest, err := c.Socketpair()
	if err != nil {
		return nil, nil, err
	}

	opt := &rpc.Options{
		BaseContext:     func() context.Context { return ctx },
		BootstrapClient: c.NewBootstrapClient(),
	}

	return rpc.NewConn(rpc.NewStreamTransport(host), opt), guest, nil
}

// socketpair creates a pair of connected Unix domain sockets for bidirectional communication
func (SocketConfig) Socketpair() (*os.File, *os.File, error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, nil, err
	}

	host := os.NewFile(uintptr(fds[0]), "host")
	guest := os.NewFile(uintptr(fds[1]), "guest")

	return host, guest, nil
}

func (c SocketConfig) NewBootstrapClient() capnp.Client {
	server := Importer_NewServer(c.Membrane)
	return capnp.NewClient(server)
}
