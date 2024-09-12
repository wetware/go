//go:generate mockgen -source=auth.go -destination=auth_mock_test.go -package=system_test AuthProvider

package system

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/record"
	"github.com/pkg/errors"
)

// AuthDomain_Nonce is the domain string used for peer records contained in an Envelope.
const AuthDomain_Nonce = "ww/auth/nonce"
const nonceSize = 20 // 160 bits

// PeerRecordEnvelopePayloadType is the type hint used to identify peer records in an Envelope.
// Defined in https://github.com/multiformats/multicodec/blob/master/table.csv
// with name "libp2p-peer-record".
var AuthDomain_PayloadType = []byte{0xbb, 0xbb} // TODO:  pick better numbers

var _ Policy = (*Terminal_login_Results)(nil)
var _ Terminal_Server = (*TerminalConfig)(nil)
var _ record.Record = (*expect)(nil)

type AuthProvider interface {
	BindPolicy(user crypto.PubKey, policy Policy) error
}

type Policy interface {
	NewStdio() (Socket, error)
	SetStdio(Socket) error
	Stdio() (Socket, error)
}

type Session struct {
	Stdio Socket_Future
}

func (s Session) Reader(ctx context.Context) io.Reader {
	return pipeReader{ReadPipe: s.Stdio.Reader()}
}

func (s Session) Writer(ctx context.Context) io.WriteCloser {
	return pipeWriter{WritePipe: s.Stdio.Writer()}
}

func (s Session) ErrWriter(ctx context.Context) io.WriteCloser {
	return pipeWriter{WritePipe: s.Stdio.Error()}
}

type TerminalConfig struct {
	Rand io.Reader
	Auth AuthProvider
}

func (t TerminalConfig) Build() Terminal {
	return Terminal_ServerToClient(t)
}

func (t TerminalConfig) Login(ctx context.Context, login Terminal_login) error {
	nonce, err := t.NewNonce()
	if err != nil {
		return err
	}

	login.Go()

	// Send the nonce over to the caller for signing.
	account := login.Args().Account()
	f, release := account.Sign(ctx, func(p Signer_sign_Params) error {
		return p.SetData(nonce)
	})
	defer release()

	// Await the signature from the caller.
	s, err := f.Struct()
	if err != nil {
		return err
	}

	envelope, err := s.RawEnvelope()
	if err != nil {
		return err
	}

	// Verify that the signature is valid for the nonce.
	user, err := t.Verify(envelope, nonce)
	if err != nil {
		return err
	}

	policy, err := login.AllocResults()
	if err != nil {
		return err
	}

	// Bind policy for the public key of the caller.
	return t.Auth.BindPolicy(user, policy)
}

func (t TerminalConfig) NewNonce() (nonce []byte, err error) {
	if t.Rand == nil {
		t.Rand = rand.Reader
	}

	nonce = make([]byte, nonceSize) // 160 bits

	var n int
	if n, err = io.ReadFull(t.Rand, nonce); err != nil {
		err = fmt.Errorf("read rand: got %d bytes: %w", n, err)
	}

	return
}

func (t TerminalConfig) Verify(sig []byte, nonce []byte) (crypto.PubKey, error) {
	e, err := record.ConsumeTypedEnvelope(sig, expect(nonce))
	if err != nil {
		return nil, err
	}

	return e.PublicKey, nil
}

type SignOnce struct {
	PrivKey crypto.PrivKey
}

func (s *SignOnce) Sign(ctx context.Context, sign Signer_sign) error {
	if s.PrivKey == nil {
		return errors.New("public key missing or revoked")
	}

	data, err := sign.Args().Data()
	if err != nil {
		return err
	}

	res, err := sign.AllocResults()
	if err != nil {
		return err
	}

	e, err := record.Seal(expect(data), s.PrivKey)
	if err != nil {
		return err
	}

	raw, err := e.Marshal()
	if err != nil {
		return err
	}

	return res.SetRawEnvelope(raw)
}

type expect []byte

// Domain is the "signature domain" used when signing and verifying a particular
// Record type. The Domain string should be unique to your Record type, and all
// instances of the Record type must have the same Domain string.
func (e expect) Domain() string {
	return AuthDomain_Nonce
}

// Codec is a binary identifier for this type of record, ideally a registered multicodec
// (see https://github.com/multiformats/multicodec).
// When a Record is put into an Envelope (see record.Seal), the Codec value will be used
// as the Envelope's PayloadType. When the Envelope is later unsealed, the PayloadType
// will be used to look up the correct Record type to unmarshal the Envelope payload into.
func (e expect) Codec() []byte {
	return AuthDomain_PayloadType
}

// MarshalRecord converts a Record instance to a []byte, so that it can be used as an
// Envelope payload.
func (e expect) MarshalRecord() ([]byte, error) {
	if len(e) != nonceSize {
		return nil, fmt.Errorf("expected 160 bit nonce, got %d", bitlen(e))
	}

	return e, nil
}

// UnmarshalRecord unmarshals a []byte payload into an instance of a particular Record type.
func (e expect) UnmarshalRecord(b []byte) error {
	if len(b) != nonceSize {
		return fmt.Errorf("expected 160 bit nonce, got %d", bitlen(b))
	}

	if copy(e, b) != nonceSize {
		panic(e) // unreachable
	}

	if bytes.Equal(e, b) {
		return nil
	}

	return fmt.Errorf("expected nonce %x, got %x", e, b)
}

func bitlen(b []byte) int {
	return len(b) * 8
}
