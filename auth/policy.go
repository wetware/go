package auth

import (
	context "context"
	"errors"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	system "github.com/wetware/go/system"
)

type Challenge func(Signer_sign_Params) error

type Policy interface {
	Bind(context.Context, Terminal_login_Results, peer.ID) error // TODO:  use another type instead of peer.ID to represent accounts
}

type SingleUser struct {
	User    crypto.PubKey
	IPFS    system.IPFS
	Exec    system.Executor
	Console *Console
}

func (policy SingleUser) Bind(ctx context.Context, env Terminal_login_Results, user peer.ID) error {
	allowed, err := peer.IDFromPublicKey(policy.User)
	if err != nil {
		return err
	}

	if user != allowed {
		return errors.New("user not allowed")
	}

	// Bind IPFS capability
	err = env.SetIpfs(policy.IPFS.AddRef())
	if err != nil {
		return err
	}

	// Bind Exec capability
	err = env.SetExec(policy.Exec.AddRef())
	if err != nil {
		return err
	}

	// Bind Console capability
	consoleClient := system.Console_ServerToClient(policy.Console)
	return env.SetConsole(consoleClient)
}

// Console implements system.Console_Server to print to an io.Writer
type Console struct {
	writer io.Writer
}

// NewConsole creates a new Console that writes to the specified io.Writer
func NewConsole(writer io.Writer) *Console {
	return &Console{writer: writer}
}

func (c *Console) Println(ctx context.Context, call system.Console_println) error {
	// Get the output data from the call
	output, err := call.Args().Output()
	if err != nil {
		return err
	}

	// Write to the writer with a newline
	bytesWritten, err := fmt.Fprintln(c.writer, output)
	if err != nil {
		return err
	}

	// Set the result (number of bytes written)
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	results.SetN(uint32(bytesWritten))

	return nil
}
