package auth

import (
	"context"
	"encoding/binary"
	"io"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/record"
)

type DefaultTerminal struct {
	Rand   io.Reader
	Policy Policy
}

func (term DefaultTerminal) Login(ctx context.Context, call Terminal_login) error {
	user := call.Args().Account()
	defer user.Release()

	var n Nonce
	f, release := user.Sign(ctx, term.NewChallenge(n[:]))
	defer release()

	res, err := f.Struct()
	if err != nil {
		return err
	}

	rawEnvelope, err := res.RawEnvelope()
	if err != nil {
		return err
	}

	// Check if signed envelope matches bytes in n.
	e, err := record.ConsumeTypedEnvelope(rawEnvelope, &n)
	if err != nil {
		return err
	}
	// validation passed

	// allocate and return schema and value
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	// TODO:  we need a separate type for accounts so that we don't
	// confuse them with actual peer.IDs.
	userID, err := peer.IDFromPublicKey(e.PublicKey)
	if err != nil {
		return err
	}

	return term.Policy.Bind(ctx, results, userID)
}

func (term DefaultTerminal) GenerateChallenge(rand io.Reader) Challenge {
	return func(call Signer_sign_Params) error {
		var n Nonce
		if err := binary.Read(rand, binary.LittleEndian, n[:]); err != nil {
			return err
		}

		return call.SetSrc(n[:])
	}
}

func (term DefaultTerminal) NewChallenge(nonce []byte) Challenge {
	return func(s Signer_sign_Params) error {
		if _, err := io.ReadFull(term.Rand, nonce); err != nil {
			return err
		}

		return s.SetSrc(nonce)
	}
}
