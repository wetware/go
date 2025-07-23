package run

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/auth"
	"github.com/wetware/go/util"
)

var env util.Env

func Command() *cli.Command {
	return &cli.Command{
		Name: "run",
		// ArgsUsage: "<source-dir>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "ipfs",
				Value: "/dns4/localhost/tcp/5001/http",
			},
			&cli.StringSliceFlag{
				Name:    "env",
				EnvVars: []string{"WW_ENV"},
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:   "membrane",
				Hidden: true,
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()

					// To anyone tempted to collapse this function by
					// refactoring `login` to consume `c`, be aware
					// that we are deliberately withholding CLI input
					// from the isolated code paths.

					t := rpc.NewStreamTransport(os.Stdin)
					defer t.Close()

					return isolate(ctx, t)
				},
			},
		},
		Before: env.Setup,
		Action: Main,
		After:  env.Teardown,
	}
}

func Main(c *cli.Context) error {
	// Set up temporary filesystem
	////
	tmpDir, err := os.MkdirTemp("", "ww-*")
	if err != nil {
		return fmt.Errorf("mkdirtemp: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			slog.ErrorContext(c.Context, "failed to remove tempdir",
				"path", tmpDir,
				"reason", err)
		}
	}()

	cellDir, err := os.MkdirTemp(tmpDir, "cmd-*")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(cellDir); err != nil {
			slog.ErrorContext(c.Context, "failed to remove cmd dir",
				"path", cellDir,
				"reason", err)
		}
	}()
	fmt.Println(cellDir)

	// Load the child's identity and initialize a one-time signer
	////
	secret, secretFile, err := bind(cellDir)
	if err != nil {
		return err
	}
	defer secretFile.Close()

	// Build the exports list.
	////

	// Set up the host capnp environment
	////
	host, guest, err := BootConfig{
		Term: auth.DefaultTerminal{
			// This is the bootstrap terminal we will provide to the guest.
			// It is important that this be the ONLY place where effectful
			// things are passed into the guest.

			Rand: rand.Reader,
			// crypto rand is an example of an effectful thing that is worth
			// passing down to the guest.

			Policy: auth.SingleUser{
				// Policy determines which bootstrap capabilities are exported
				// to the guest.
				//
				// auth.SingleUser implements Policy.  It provides the node
				// only to guests that log in as `user`.

				User:   secret.GetPublic(), // user to allow
				Export: auth.SchemaProvider(capnp.ErrorClient(errors.New("SchemaProvider::not implemented"))),
			},
		},
	}.Boot()
	if err != nil {
		return fmt.Errorf("bind: %w", err)
	}
	defer host.Close()
	defer guest.Close()

	// Set up the guest
	////

	// Step 1: rewrite "ww run foo [...]" ==> "ww run isolate foo [...]"
	args := append([]string{"run", "membrane"}, c.Args().Tail()...)

	// Step 2:  run the subcommand
	cmd := exec.CommandContext(c.Context, "ww", args...)
	cmd.Dir = cellDir
	cmd.Env = c.StringSlice("env")
	cmd.Stdin = guest
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = sysProcAttr(cellDir)  // TODO:  identity file is visible here; do we want that?
	cmd.ExtraFiles = []*os.File{secretFile} // fd 3 in child

	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Cancel()

	// client := host.Bootstrap(c.Context)
	// defer client.Release()

	// release := system.Export{
	// 	Proto:  "/ww/0.1.0",
	// 	Client: client, // system.Executor
	// 	// TODO:  schema; export via the /schema subprotocol
	// }.Bind(c.Context, env.Host)
	// defer release()

	return cmd.Wait()
}

func bind(cellDir string) (crypto.PrivKey, *os.File, error) {
	identity, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	accountSecretBytes, err := crypto.MarshalPrivateKey(identity)
	if err != nil {
		return nil, nil, err
	}

	// Create temporary identity file
	identityPath := cellDir + "/identity"
	fileWriter, err := os.OpenFile(identityPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, nil, err
	}
	defer fileWriter.Close()

	_, err = io.Copy(fileWriter, bytes.NewReader(accountSecretBytes))
	if err != nil { // always close identityFile
		return nil, nil, err
	}

	// Reopen for reading to pass to child
	f, err := os.Open(identityPath)
	return identity, f, err
}

type BootConfig struct {
	Term auth.Terminal_Server
}

func (l BootConfig) Boot() (host *rpc.Conn, guest *os.File, err error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		err = fmt.Errorf("socketpair: %w", err)
	} else {
		hostf := os.NewFile(uintptr(fds[0]), "host")
		host = rpc.NewConn(rpc.NewStreamTransport(hostf), &rpc.Options{
			// boot.Config exposes a Client() method that we  can call in-
			// line.
			BootstrapClient: capnp.Client(auth.Terminal_ServerToClient(l.Term)),
		})

		guest = os.NewFile(uintptr(fds[1]), "guest")
	}

	return
}

func isolate(ctx context.Context, t rpc.Transport) error {
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

	conn := rpc.NewConn(t, nil)
	defer conn.Close()

	term := auth.Terminal(conn.Bootstrap(ctx))
	f, release := term.Login(ctx, user(id))
	defer release()

	fschema, release := f.Session().Schema(ctx, nil)
	defer release()

	fmt.Println(fschema.Schema())

	return nil
}

func user(privKey crypto.PrivKey) func(auth.Terminal_login_Params) error {
	return func(call auth.Terminal_login_Params) error {
		server := &auth.SignOnce{PrivKey: privKey}
		client := auth.Signer_ServerToClient(server)
		return call.SetAccount(client)
	}
}
