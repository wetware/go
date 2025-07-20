//go:generate capnp compile -I.. -I$GOPATH/src/capnproto.org/go/capnp/std -ogo boot.capnp

package boot

import (
	"fmt"
	"os"
	"syscall"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/wetware/go/auth"
)

type DefaultLoader[T auth.Terminal_Server] struct {
	Term T
}

func (c DefaultLoader[T]) Boot() (host *rpc.Conn, guest *os.File, err error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		err = fmt.Errorf("socketpair: %w", err)
	} else {
		hostf := os.NewFile(uintptr(fds[0]), "host")
		host = rpc.NewConn(rpc.NewStreamTransport(hostf), &rpc.Options{
			// boot.Config exposes a Client() method that we  can call in-
			// line.
			BootstrapClient: capnp.Client(auth.Terminal_ServerToClient(c.Term)),
		})

		guest = os.NewFile(uintptr(fds[1]), "guest")
	}

	return
}
