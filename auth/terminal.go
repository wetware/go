package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/record"
	"github.com/pkg/errors"
)

type TerminalConfig struct {
	Rand io.Reader
	Auth Provider
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

// Session is a login context for a terminal.  It binds the
// guest's standard I/O to the host runtime.  For guest code,
// the contract is as follows:
//
//   - The guest's standard input is produced by a Cap'n Proto
//     RPC connection.
//
//   - The guest's standard output is consumed by a Cap'n Proto
//     RPC connection.
//
//   - The guest's standard error is consumed and processed into
//     wetware event logs.
type Session struct {
	// Proc exposes standard I/O to a guest process.
	Proc interface {
		Reader() ReadPipe  // guest's stdin
		Writer() WritePipe // guest's stdout
		Error() WritePipe  // guest's stderr
	}

	// Sock provides an abstraction over the bridging of
	// Cap'n Proto RPC calls and Go's io.Reader/Writer.
	Sock interface {
		Bind(context.Context, WritePipe) io.WriteCloser
		Connect(context.Context, ReadPipe) io.Reader
	}
}

func (s Session) Reader(ctx context.Context) io.Reader {
	rpipe := s.Proc.Reader()
	return s.Sock.Connect(ctx, rpipe)
}

func (s Session) Writer(ctx context.Context) io.WriteCloser {
	wpipe := s.Proc.Writer()
	return s.Sock.Bind(ctx, wpipe)
}

func (s Session) ErrWriter(ctx context.Context) io.WriteCloser {
	wpipe := s.Proc.Error()
	return s.Sock.Bind(ctx, wpipe)
}
