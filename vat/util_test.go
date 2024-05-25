package vat_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/wetware/go/system"
	"github.com/wetware/go/vat"
)

func TestProtoFromModule(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	wasi.MustInstantiate(ctx, r)

	sys, err := system.Builder{}.Instantiate(ctx, r)
	require.NoError(t, err)
	defer sys.Close(ctx)

	b, err := os.ReadFile("testdata/socket/main.wasm")
	require.NoError(t, err)

	cm, err := r.CompileModule(ctx, b)
	require.NoError(t, err)
	defer cm.Close(ctx)

	mod, err := r.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
		WithName("test"))
	require.NoError(t, err)
	defer mod.Close(ctx)

	require.Equal(t,
		filepath.Join(vat.Proto, "test"),
		string(vat.ProtoFromModule(mod)))
}
