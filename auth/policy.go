package auth

import (
	context "context"
	"errors"
	"fmt"
	"io"

	capnp "capnproto.org/go/capnp/v3"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/record"
)

type Challenge func(Signer_sign_Params) error

type Policy interface {
	Bind(context.Context, Terminal_login_Results, peer.ID) error // TODO:  use another type instead of peer.ID to represent accounts
}

type SingleUser[T ~capnp.StructKind] struct {
	User Signer
	Rand io.Reader
	Env  T
}

func (policy SingleUser[T]) Bind(ctx context.Context, sess Terminal_login_Results, user peer.ID) error {
	var n Nonce
	f, release := policy.User.Sign(ctx, func(call Signer_sign_Params) error {
		if n, err := io.ReadFull(policy.Rand, n[:]); err != nil {
			return fmt.Errorf("byte %d: %w", n, err)
		}
		return call.SetSrc(n[:])
	})
	defer release()

	res, err := f.Struct()
	if err != nil {
		return err
	}

	rawEnvelope, err := res.RawEnvelope()
	if err != nil {
		return err
	}

	e, err := record.ConsumeTypedEnvelope(rawEnvelope, &n)
	if err != nil {
		return err
	}

	if !user.MatchesPublicKey(e.PublicKey) {
		return errors.New("rejected: users don't match")
	}

	// if err := sess.SetType(policy.Type); err != nil {
	// 	return err
	// }

	// client := capnp.NewClient(Env_NewServer(policy.Env))
	// if err := sess.SetNode(client); err != nil {
	// 	defer client.Release()
	// 	return err
	// }

	return nil
}
