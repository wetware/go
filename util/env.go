package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"capnproto.org/go/capnp/v3/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/auth"
)

func ExpandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

type Env struct {
	IPFS iface.CoreAPI
	Host host.Host
}

func (env *Env) Setup(c *cli.Context) (err error) {
	for _, bind := range []bindFunc{
		func(c *cli.Context) (err error) {
			env.IPFS, err = LoadIPFSFromName(c.String("ipfs"))
			return
		},
		// func(ctx *cli.Context) (err error) {
		// 	env.Host, err = LoadHost(env.IPFS)
		// 	return
		// },
	} {
		if err = bind(c); bind != nil {
			break
		}
	}

	return
}

type bindFunc func(*cli.Context) (err error)

func (env *Env) Teardown(c *cli.Context) (err error) {
	if env.Host != nil {
		err = env.Host.Close()
	}

	return
}

type CellFunc func(context.Context, auth.Terminal_login_Results) error

func Isolate(ctx context.Context, isolate CellFunc) (err error) {
	// Get host socket from environment (fd 3)
	host := os.NewFile(3, "host")
	defer host.Close()

	identity := os.NewFile(4, "identity")
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

	conn := rpc.NewConn(rpc.NewStreamTransport(host), nil)
	defer conn.Close()

	term := auth.Terminal(conn.Bootstrap(ctx))
	f, release := term.Login(ctx, user(id))
	defer release()

	session, err := f.Struct()
	if err != nil {
		return err
	}

	return isolate(ctx, session)
}

func user(privKey crypto.PrivKey) func(auth.Terminal_login_Params) error {
	return func(call auth.Terminal_login_Params) error {
		server := &auth.SignOnce{PrivKey: privKey}
		client := auth.Signer_ServerToClient(server)
		return call.SetAccount(client)
	}
}
