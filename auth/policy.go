package auth

import (
	context "context"
	"errors"

	iface "github.com/ipfs/kubo/core/coreiface"
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
	IPFS iface.CoreAPI
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
	provider := &system.IPFS_Provider{API: policy.IPFS}
	ipfs := system.IPFS_ServerToClient(provider)
	return env.SetIpfs(ipfs)
}
