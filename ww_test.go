package ww_test

import (
	"bytes"
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/ipfs/kubo/client/rpc"
	"github.com/libp2p/go-libp2p"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	ww "github.com/wetware/go"
)

//go:embed testdata/main.wasm
var fileContent []byte

func TestEcho(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	ipfs, err := rpc.NewLocalApi()
	require.NoError(t, err)

	h, err := libp2p.New()
	require.NoError(t, err)
	defer h.Close()

	stdin := bytes.NewBufferString(`test`)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	err = ww.Env{
		IPFS: ipfs,
		Host: h,
		Boot: bytecode(fileContent),
		WASM: wazero.NewRuntimeConfig().
			// WithDebugInfoEnabled(true).
			WithCloseOnContextDone(true),
		Module: wazero.NewModuleConfig().
			WithArgs("/testdata/main.wasm").
			WithStdin(stdin).
			WithStdout(stdout).
			WithStderr(stderr).
			WithFSConfig(wazero.NewFSConfig().
				WithDirMount("testdata", "/testdata/")),
	}.Serve(ctx)

	require.NoError(t, err, "server failed")
	require.Equal(t, "test", stdout.String(), "unexpected output")
}

type bytecode []byte

func (b bytecode) Load(context.Context) ([]byte, error) {
	return b, nil
}
