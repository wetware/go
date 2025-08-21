package main

import (
	"context"
	"fmt"
	"os"

	"capnproto.org/go/capnp/v3/rpc"
	export_cap "github.com/wetware/go/examples/export/cap"
	"github.com/wetware/go/system"
)

func fail(error string) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", error)
	os.Exit(1)
}

func main() {
	ctx := context.Background()

	// Check if the bootstrap file descriptor exists
	bootstrapFile := os.NewFile(system.BOOTSTRAP_FD, "host")
	if bootstrapFile == nil {
		fail("ERROR: Failed to create bootstrap file descriptor\n")
	}

	conn := rpc.NewConn(rpc.NewStreamTransport(bootstrapFile), &rpc.Options{
		BaseContext: func() context.Context { return ctx },
		// BootstrapClient: export(),
	})
	defer conn.Close()

	client := conn.Bootstrap(ctx)
	defer client.Release()

	greeter := export_cap.Greeter(client)
	f, release := greeter.Greet(ctx, func(params export_cap.Greeter_greet_Params) error {
		return params.SetName("Import Example")
	})
	defer release()

	// Wait for the greeting to complete
	<-f.Done()

	res, err := f.Struct()
	if err != nil {
		fail(fmt.Sprintf("ERROR: Failed to get greeting: %v\n", err))
	}

	greeting, err := res.Greeting()
	if err != nil {
		fail(fmt.Sprintf("ERROR: Failed to get greeting: %v\n", err))
	}

	fmt.Println(greeting)
}
