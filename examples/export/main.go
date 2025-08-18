//go:generate capnp compile -I$GOPATH/src/capnproto.org/go/capnp/std -ogo export.capnp

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/wetware/go/system"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM)
	defer cancel()

	// Open the bootstrap socket that was passed in from the host.
	sock := os.NewFile(system.BOOTSTRAP_FD, "host")
	defer sock.Close()

	// Wrap the socket in a Cap'n Proto connection.
	conn := rpc.NewConn(rpc.NewStreamTransport(sock), &rpc.Options{
		BaseContext:     func() context.Context { return ctx },
		BootstrapClient: export(),
	})
	defer conn.Close()

	// Get the bootstrap client from the host.
	conn.Bootstrap(ctx)
}

func export() capnp.Client {
	server := Greeter_NewServer(greeter{})
	return capnp.NewClient(server)
}

type greeter struct{}

func (greeter) Greet(_ context.Context, call Greeter_greet) error {
	name, err := call.Args().Name()
	if err != nil {
		return err
	}

	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	return res.SetGreeting(fmt.Sprintf("Hello, %s! 👋", name))
}
