package auth

import (
	context "context"
	"errors"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	system "github.com/wetware/go/system"
)

type Challenge func(Signer_sign_Params) error

type Policy interface {
	Bind(context.Context, Terminal_login_Results, peer.ID) error // TODO:  use another type instead of peer.ID to represent accounts
}

type SingleUser struct {
	User crypto.PubKey
	IPFS system.IPFS
	Exec system.Executor
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

	return env.SetExec(policy.Exec.AddRef())
}
