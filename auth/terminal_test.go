package auth_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/auth"
)

func TestTerminal_Login(t *testing.T) {
	t.Parallel()

	rnd := rand.New(rand.NewSource(42))
	priv, pub, err := crypto.GenerateECDSAKeyPair(rnd)
	require.NoError(t, err)

	term := auth.Terminal_ServerToClient(auth.TerminalConfig{
		Rand: rnd,
		Auth: expectUser{PubKey: pub},
	})
	defer term.Release()

	t.Run("ShouldAccept", func(t *testing.T) {
		f, release := term.Login(context.TODO(), func(login auth.Terminal_login_Params) error {
			return login.SetAccount(auth.Signer_ServerToClient(&auth.SignOnce{
				PrivKey: priv,
			}))
		})
		defer release()

		res, err := f.Struct()
		require.NoError(t, err)

		_, err = res.Stdio()
		require.NoError(t, err)
	})

	t.Run("ShouldReject", func(t *testing.T) {
		rnd := rand.New(rand.NewSource(42))
		newPriv, _, err := crypto.GenerateECDSAKeyPair(rnd)
		require.NoError(t, err)

		f, release := term.Login(context.TODO(), func(login auth.Terminal_login_Params) error {
			return login.SetAccount(auth.Signer_ServerToClient(&auth.SignOnce{
				PrivKey: newPriv,
			}))
		})
		defer release()

		_, err = f.Struct()
		require.NoError(t, err)
	})
}

type expectUser struct {
	crypto.PubKey
}

func (e expectUser) BindPolicy(user crypto.PubKey, policy auth.Policy) error {
	if user.Equals(e.PubKey) {
		return nil
	}

	return errors.New("pubkeys don't match")
}
