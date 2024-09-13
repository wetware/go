package auth_test

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	auth "github.com/wetware/go/auth"
)

func TestAuthProvider(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h, err := libp2p.New(
		libp2p.NoTransports,
		libp2p.NoListenAddrs)
	require.NoError(t, err)

	user := privkey(h)
	account := auth.Signer_ServerToClient(&auth.SignOnce{
		PrivKey: user,
	})

	var called bool
	p := NewMockProvider(ctrl)
	p.EXPECT().
		BindPolicy(gomock.Any(), gomock.Any()).
		DoAndReturn(func(user crypto.PubKey, policy auth.Terminal_login_Results) error {
			called = true
			return nil
		}).
		Times(1)

	// Create a terminal server.
	terminal := auth.TerminalConfig{
		Rand: rand.Reader,
		Auth: p,
	}.Build()

	f, release := terminal.Login(context.Background(), func(p auth.Terminal_login_Params) error {
		return p.SetAccount(account)
	})
	defer release()

	_, err = f.Struct()
	require.NoError(t, err)
	assert.True(t, called, "AuthProvider was not called")
}

func privkey(h host.Host) crypto.PrivKey {
	return h.Peerstore().PrivKey(h.ID())
}