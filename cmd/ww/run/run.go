package run

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"capnproto.org/go/capnp/v3/std/capnp/schema"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/auth"
	"github.com/wetware/go/boot"
)

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
					// refactoring `isolate` to consume `c`, be aware
					// that we are deliberately withholding CLI input
					// from the isolated code paths.
					return isolate(ctx)
				},
			},
		},
		Before: setup,
		Action: Main,
		// After: teardown,
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

	cellDir, err := os.MkdirTemp(tmpDir, "cell-*")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(cellDir); err != nil {
			slog.ErrorContext(c.Context, "failed to remove cell dir",
				"path", cellDir,
				"reason", err)
		}
	}()
	fmt.Println(cellDir)

	// Set up Host
	////
	secret, secretFile, err := bind(cellDir)
	if err != nil {
		return err
	}

	signer := &auth.SignOnce{PrivKey: secret}
	user := auth.Signer_ServerToClient(signer)

	// Step 2:  set up the host capnp environment

	host, guest, err := boot.DefaultLoader[auth.DefaultTerminal]{
		Term: auth.DefaultTerminal{
			// This is the bootstrap terminal we will provide to the guest.
			// It is important that this be the ONLY place where effectful
			// things are passed into the guest.

			Rand: rand.Reader,
			// crypto rand is an example of an effectful thing that is worth
			// passing down to the guest.

			Policy: auth.SingleUser[Cell]{
				// Policy determines which bootstrap capability is exported
				// to the guest.
				//
				// auth.SingleUser implements Policy.  It provides the node
				// only to guests that log in as `user`.

				User: user,
				Rand: rand.Reader,
				Node: Cell{Rand: rand.Reader, IPFS: env.IPFS},
				Type: schema.Node{}}, // FIXME:  empty schema.Node{}
		}}.Boot()
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
	cmd.ExtraFiles = []*os.File{secretFile} // fd 3 in child

	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Cancel()

	/*		SCRATCH		*/

	// // Run host loop
	// ////
	// client := conn.Bootstrap(c.Context)
	// defer client.Release()

	// TODO:
	// for each stream that comes into libp2p
	// export a `client` to the remote endpoint
	// hold until done

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
