package proc_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"capnproto.org/go/capnp/v3"
	"github.com/wetware/go/proc"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/stretchr/testify/require"
)

func TestProc_echo(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	b, err := os.ReadFile("../examples/echo/main.wasm")
	require.NoError(t, err, "failed to open file")
	require.NotEmpty(t, b, "empty file")

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	cm, err := r.CompileModule(ctx, b)
	require.NoError(t, err, "failed to compile module")
	defer cm.Close(ctx)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	p, err := proc.Config{
		Stdout: stdout,
		Stderr: stderr,
	}.Instantiate(ctx, r, cm)
	require.NoError(t, err, "failed to instantiate process")
	require.NotNil(t, p, "should return *proc.P")
	defer p.Close(ctx)

	msg, seg := capnp.NewSingleSegmentMessage(nil)
	call, err := proc.NewRootMethodCall(seg)
	require.NoError(t, err)
	err = call.SetName("echo")
	require.NoError(t, err)
	err = call.SetCallData([]byte("Hello, Wetware!"))
	require.NoError(t, err)
	defer msg.Release()

	err = p.Deliver(ctx, call)
	require.NoError(t, err, "delivery failed")

	require.Equal(t, "", stderr.String(), "unexpected error")
	require.Equal(t, "Hello, Wetware!", stdout.String(), "unexpected output")
}

// TestProc_echo_repeated_calls asserts that guest export can be
// called repeatedly.  The guest code is stateless.
func TestProc_echo_repeated_calls(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	b, err := os.ReadFile("../examples/echo/main.wasm")
	require.NoError(t, err, "failed to open file")
	require.NotEmpty(t, b, "empty file")

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	cm, err := r.CompileModule(ctx, b)
	require.NoError(t, err, "failed to compile module")
	defer cm.Close(ctx)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	p, err := proc.Config{
		Stdout: stdout,
		Stderr: stderr,
	}.Instantiate(ctx, r, cm)
	require.NoError(t, err, "failed to instantiate process")
	require.NotNil(t, p, "should return *proc.P")
	defer p.Close(ctx)

	for i := 0; i < 10; i++ {
		func(t *testing.T, i int) {
			defer stdout.Reset()
			defer stderr.Reset()

			msg, seg := capnp.NewSingleSegmentMessage(nil)
			call, err := proc.NewRootMethodCall(seg)
			require.NoError(t, err)
			err = call.SetName("echo")
			require.NoError(t, err)
			err = call.SetCallData([]byte("Hello, Wetware!"))
			require.NoError(t, err)
			defer msg.Release()

			err = p.Deliver(ctx, call)
			require.NoError(t, err, "delivery failed")

			require.Equal(t, "", stderr.String(), "unexpected error")
			require.Equal(t, "Hello, Wetware!", stdout.String(), "unexpected output")
		}(t, i)
	}
}
