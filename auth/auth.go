//go:generate capnp compile -I../api -I$GOPATH/src/capnproto.org/go/capnp/std -ogo auth.capnp

package auth

import (
	context "context"
	"errors"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/record"
)

type Nonce [16]byte

func (Nonce) Domain() string {
	return "ww.auth"
}

func (Nonce) Codec() []byte {
	return []byte{0xde, 0xea} // TODO:  pick a good value for this
}

func (n Nonce) MarshalRecord() ([]byte, error) {
	return n[:], nil
}

func (n *Nonce) UnmarshalRecord(buf []byte) error {
	if len(buf) != len(n) {
		return fmt.Errorf("invalid nonce size: %d", len(buf))
	}

	var candidate Nonce
	copy(candidate[:], buf)

	// Only copy over if the signature matches the nonce
	if candidate != *n {
		return fmt.Errorf("nonce signature mismatch")
	}

	return nil
}

type SignOnce struct {
	sync.Once
	PrivKey crypto.PrivKey
}

func (once *SignOnce) Sign(ctx context.Context, call Signer_sign) (err error) {
	err = errors.New("already called")
	once.Do(func() {
		err = once.bind(call)
	})
	return
}

func (once *SignOnce) bind(call Signer_sign) error {
	src, err := call.Args().Src()
	if err != nil {
		return err
	}

	var n Nonce
	if u := copy(n[:], src); u != len(n) {
		return fmt.Errorf("invalid nonce len: %d", u)
	}

	e, err := record.Seal(&n, once.PrivKey)
	if err != nil {
		return err
	}

	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	envelopePayload, err := e.Marshal()
	if err != nil {
		return err
	}

	return res.SetRawEnvelope(envelopePayload)
}
