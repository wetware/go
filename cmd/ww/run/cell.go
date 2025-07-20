package run

import (
	"fmt"
	"io"
	"os"

	"capnproto.org/go/capnp/v3/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/auth"
)

func isolate(c *cli.Context) error {
	identity := os.NewFile(3, "identity")
	defer identity.Close()

	raw, err := io.ReadAll(identity)
	if err != nil {
		return fmt.Errorf("read identity: %w", err)
	}
	identity.Close()

	// Local identity.  Host will accept signed messages with this key.
	id, err := crypto.UnmarshalPrivateKey(raw)
	if err != nil {
		return fmt.Errorf("unmarshal identity: %w", err)
	}

	conn := rpc.NewConn(rpc.NewStreamTransport(os.Stdin), &rpc.Options{
		// ...
	})
	defer conn.Close()

	term := auth.Terminal(conn.Bootstrap(c.Context))
	f, release := term.Login(c.Context, user(id))
	defer release()

	// Resolve the future to get the actual results
	results, err := f.Struct()
	if err != nil {
		return fmt.Errorf("failed to resolve login results: %w", err)
	}

	// Get the type and then grab the node itself
	schema, err := results.Type()
	if err != nil {
		return err
	}
	node := results.Node()
	if !node.IsValid() {
		debug := node.Snapshot().Brand().Value
		return fmt.Errorf("invalid node: %v", debug)
	}

	// Convert to the expected type

	fmt.Println(node)
	fmt.Println(schema)
	return nil

	// for c.Context.Err() == nil {
	// 	// read input value

	// 	// evaluate

	// 	// print
	// }

	// return c.Context.Err()
}

func user(privKey crypto.PrivKey) func(auth.Terminal_login_Params) error {
	return func(call auth.Terminal_login_Params) error {
		server := &auth.SignOnce{PrivKey: privKey}
		client := auth.Signer_ServerToClient(server)
		return call.SetAccount(client)
	}
}

// var _ boot.Env_Server = (*Cell)(nil)

type Cell struct {
	Rand io.Reader
	IPFS iface.CoreAPI
}
