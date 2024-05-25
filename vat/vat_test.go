package vat_test

import (
	"context"
	"crypto/rand"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/wetware/go/system"
	test_libp2p "github.com/wetware/go/test/libp2p"
	"github.com/wetware/go/vat"
)

func TestNetConfig(t *testing.T) {
	t.Parallel()

	t.Skip("XXX")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).
		WithDebugInfoEnabled(false))
	defer r.Close(ctx)

	cl, err := wasi.Instantiate(ctx, r)
	require.NoError(t, err)
	defer cl.Close(ctx)

	sys, err := system.Builder{
		// Host:    c.Host,
		// IPFS:    c.IPFS,
	}.Instantiate(ctx, r)
	require.NoError(t, err)
	defer sys.Close(ctx)

	b, err := os.ReadFile("testdata/socket/main.wasm")
	require.NoError(t, err)

	cm, err := r.CompileModule(ctx, b)
	require.NoError(t, err)
	defer cm.Close(ctx)

	ch := make(chan []byte, 1)
	defer close(ch)
	ctx = system.WithMailbox(ctx, ch)

	mod, err := r.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
		// WithName().
		// WithArgs().
		// WithEnv().
		WithRandSource(rand.Reader).
		// WithFS().
		// WithFSConfig().
		// WithStartFunctions(). // remove _start so that we can call it later
		WithStdin(sys.Stdin()).
		WithStdout(os.Stdout). // FIXME
		WithStderr(os.Stderr). // FIXME
		WithSysNanotime())
	require.NoError(t, err)
	defer mod.Close(ctx)

	h := test_libp2p.NewMockHost(ctrl)

	net := vat.NetConfig{
		Host: h,
	}.Build(ctx)

	require.NotZero(t, net.DialTimeout,
		"should initialize dial_timeout parameter")

}
