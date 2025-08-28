package main

import (
	"context"
	"fmt"
	"os"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"

	export_cap "github.com/wetware/go/examples/export/cap"
)

func fail(error string) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", error)
	os.Exit(1)
}

func main() {
	ctx := context.Background()

	if len(os.Args) != 3 {
		fmt.Println(`Usage:
			./<executable> <export peer id> <export peer maddr>
Example:
			./<executable> id=12D3KooWMh9743yKf2h3ZrjAGVtVBmDqyqnghxEoyqz3wSiPfv5e /ip4/127.0.0.1/tcp/2020`)
	}

	remote, remoteAddr := os.Args[1], os.Args[2]

	host, err := libp2p.New()
	if err != nil {
		fail(fmt.Sprintf("ERROR: Failed to create libp2p host: %v\n", err))
	}
	defer host.Close()

	addr, err := ma.NewMultiaddr(remoteAddr)
	if err != nil {
		fail(err.Error())
	}

	id, err := peer.Decode(remote)
	if err != nil {
		fail(err.Error())
	}

	err = host.Connect(ctx, peer.AddrInfo{
		ID:    id,
		Addrs: []ma.Multiaddr{addr},
	})
	if err != nil {
		fail(err.Error())
	}

	s, err := host.NewStream(ctx, id, protocol.ID("/ww/0.1.0"))
	if err != nil {
		fail(err.Error())
	}
	defer s.Close()

	conn := rpc.NewConn(rpc.NewPackedStreamTransport(s), &rpc.Options{
		BaseContext: func() context.Context { return ctx },
	})

	client := conn.Bootstrap(ctx)
	defer client.Release()

	if err = client.Resolve(ctx); err != nil {
		fail(err.Error())
	}

	greeter := export_cap.Greeter(client)
	f, release := greeter.Greet(ctx, func(params export_cap.Greeter_greet_Params) error {
		return params.SetName("Import Example")
	})
	defer release()

	// Wait for the greeting to complete
	<-f.Done()

	res, err := f.Struct()
	if err != nil {
		fail(fmt.Sprintf("ERROR: greet failed: %v\n", err))
	}

	greeting, err := res.Greeting()
	if err != nil {
		fail(fmt.Sprintf("ERROR: greet failed: %v\n", err))
	}

	fmt.Println(greeting)
}
