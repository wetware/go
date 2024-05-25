package system_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	wasi "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/wetware/go/system"
)

func TestSystemSocket(t *testing.T) {
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

	ch := make(chan []byte, 1)
	defer close(ch)
	ctx = system.WithMailbox(ctx, ch)

	mod, err := r.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
		WithStdin(sys.Stdin()).
		WithStdout(os.Stdout)) // support printf debugging in guest code
	require.NoError(t, err)
	defer mod.Close(ctx)

	client := sys.Bind(mod) // bind the guest module to the system socket
	defer client.Release()

	f, release := system.Proc(client).Handle(ctx, func(h system.Proc_handle_Params) error {
		return h.SetEvent([]byte("Hello, Wetware!"))
	})
	defer release()

	_, err = f.Struct()
	require.NoError(t, err)

	require.Equal(t, "Hello, Wetware!", string(<-ch))
}
