package auth

import (
	context "context"
	"errors"

	schema "capnproto.org/go/capnp/v3/std/capnp/schema"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Challenge func(Signer_sign_Params) error

type Policy interface {
	Bind(context.Context, Terminal_login_Results, peer.ID) error // TODO:  use another type instead of peer.ID to represent accounts
}

type SingleUser struct {
	User   crypto.PubKey
	Schema schema.Node
}

func (policy SingleUser) Bind(ctx context.Context, env Terminal_login_Results, user peer.ID) error {
	allowed, err := peer.IDFromPublicKey(policy.User)
	if err != nil {
		return err
	}

	if user != allowed {
		return errors.New("user not allowed")
	}

	// TODO:  as we add more fields to Terminal_login_Results, we'll need to populate them here.
	return env.SetSchema(policy.Schema)
}
