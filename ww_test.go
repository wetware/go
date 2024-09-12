package ww_test

import (
	"bytes"
	context "context"
	"io"
	"testing"

	"github.com/ipfs/boxo/path"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero/sys"
	ww "github.com/wetware/go"
)

func TestService(t *testing.T) {
	t.Parallel()

	root, err := path.NewPath("/ipfs/QmRecDLNaESeNY3oUFYZKK9ftdANBB8kuLaMdAXMD43yon")
	require.NoError(t, err)

	ipfs, err := rpc.NewLocalApi()
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	cluster := ww.Config{
		NS:   root.String(),
		IPFS: ipfs,
		// IO:   system.Streams{Writer: nopCloser{buf}},
	}.Build()

	err = cluster.Serve(context.Background())
	status := err.(*sys.ExitError).ExitCode()
	require.Zero(t, status)

	// Check that main.wasm wrote what we expect.
	require.Equal(t, "test", buf.String())
}

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }
