package boot_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/boot"
)

func TestFS(t *testing.T) {
	t.Parallel()

	d := mockDiscoverer{
		newMockAddrInfo(),
		newMockAddrInfo()}
	bootFS := boot.FS{Discoverer: d}

	f, err := bootFS.Open("test")
	require.NoError(t, err)

	var info peer.AddrInfo
	var i int
	for dec := json.NewDecoder(f); dec.More(); i++ {
		err = dec.Decode(&info)
		require.NoError(t, err)
		require.Equal(t, d[i].ID, info.ID)
	}
}

type mockDiscoverer []peer.AddrInfo

func (m mockDiscoverer) FindPeers(ctx context.Context, ns string, _ ...discovery.Option) (<-chan peer.AddrInfo, error) {
	out := make(chan peer.AddrInfo, len(m))
	defer close(out)

	for _, info := range m {
		out <- info
	}

	return out, ctx.Err()
}

func newMockAddrInfo() peer.AddrInfo {
	pk, _, err := crypto.GenerateECDSAKeyPair(rand.Reader)
	if err != nil {
		panic(err)
	}

	id, err := peer.IDFromPrivateKey(pk)
	if err != nil {
		panic(err)
	}

	return peer.AddrInfo{
		ID: id,
	}
}
