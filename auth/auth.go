package auth

import "github.com/libp2p/go-libp2p/core/crypto"

var _ Policy = (*Terminal_login_Results)(nil)
var _ Terminal_Server = (*TerminalConfig)(nil)

type Policy interface {
	NewStdio() (Socket, error)
	SetStdio(Socket) error
	Stdio() (Socket, error)
}

type Provider interface {
	BindPolicy(user crypto.PubKey, policy Policy) error
}

type Provide struct {
	// ...
}

func (p Provide) BindPolicy(user crypto.PubKey, policy Policy) error {
	// ...
	return nil
}
