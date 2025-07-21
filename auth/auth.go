//go:generate capnp compile -I../api -I$GOPATH/src/capnproto.org/go/capnp/std -ogo auth.capnp

package auth

import (
	context "context"
	"errors"
	"fmt"
	"io"
	"sync"

	capnp "capnproto.org/go/capnp/v3"
	schema "capnproto.org/go/capnp/v3/std/capnp/schema"
	iface "github.com/ipfs/kubo/core/coreiface"
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

type SessionConfig struct {
	Rand io.Reader
	IPFS iface.CoreAPI
}

func (conf SessionConfig) Must() Env {
	env, err := conf.New()
	if err != nil {
		panic(err)
	}

	return env
}

func (conf SessionConfig) New() (Env, error) {
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return Env{}, err
	}

	env, err := NewRootEnv(seg)
	if err != nil {
		return Env{}, err
	}

	if err := env.SetSchema(schema.Node{ /* FIXME */ }); err != nil {
		return Env{}, err
	}

	return env, err
}
