package glia_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/wetware/go/glia"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
	test_libp2p "github.com/wetware/go/test/libp2p"
	gomock "go.uber.org/mock/gomock"
)

var _ glia.Proc = (*proc.P)(nil)

func TestP2P(t *testing.T) {
	t.Parallel()

	// Set up our mocking infrastructure.  We'll be mocking
	// three major interfaces:
	//  1. host.Host
	//  2. glia.Router
	//  3. glia.Proc
	////
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// Initialize libp2p mocks.
	h := test_libp2p.NewMockHost(ctrl)
	id := mkPeerID(t)
	h.EXPECT().
		ID().
		Return(id).
		AnyTimes()
	env := &system.Env{
		Host: h,
	}

	r := wazero.NewRuntimeWithConfig(context.TODO(), wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	bytecode, err := os.ReadFile("../examples/echo/main.wasm")
	require.NoError(t, err)

	cm, err := r.CompileModule(ctx, bytecode)
	require.NoError(t, err)
	defer cm.Close(ctx)

	p, err := proc.Command{
		PID: proc.NewPID(),
		// Args: ,
		// Env: ,
		Stderr: os.Stderr,
		// FS: ,
	}.Instantiate(ctx, r, cm)
	require.NoError(t, err)
	defer p.Close(ctx)

	// Initialize mock router and P2P instance
	router := &mockRouter{P: p}
	p2p := &glia.P2P{
		Env:    env,
		Host:   h,
		Router: router,
	}

	s := NewMockStream(ctrl)
	s.EXPECT().
		Destination().
		Return(id.String()).
		Times(1)
	s.EXPECT().
		ProcID().
		Return(p.String()).
		AnyTimes()
	s.EXPECT().
		MethodName().
		Return("echo").
		AnyTimes()
	read1 := s.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(b []byte) (int, error) {
			const msg = "Hello, Wetware!"
			if n := copy(b, msg); n <= len(msg) {
				return n, nil
			}
			panic("overflow")
		}).
		MaxTimes(1)
	s.EXPECT().
		Read(gomock.Any()).
		After(read1).
		DoAndReturn(func([]byte) (int, error) {
			return 0, io.EOF
		}).
		Times(1)
	got := &bytes.Buffer{}
	s.EXPECT().
		Write(gomock.Any()).
		After(read1).
		DoAndReturn(func(b []byte) (int, error) {
			return got.Write(b)
		}).
		Times(1)
	s.EXPECT().
		Close().
		Return(nil).
		After(read1).
		Times(1)

	err = p2p.ServeStream(ctx, s)
	require.NoError(t, err)
	require.Equal(t, "Hello, Wetware!", got.String())
}

type mockRouter struct {
	P *proc.P
}

func (r mockRouter) GetProc(pid string) (system.Proc, error) {
	if r.P.String() == pid {
		return r.P, nil
	}

	return nil, fmt.Errorf("mockRouter: %s != %s", r.P.String(), pid)
}

func TestP2PStream(t *testing.T) {
	t.Parallel()

	proto := "12D3KooWFYcCMuKujeeDDPqnH6yHeVrnXPaCjfbYVmQ9fHxfRDtA/Wt9hMLbqHmNuCsvqCW8AuKxUjwL/echo"

	s := test_libp2p.NewMockStream(gomock.NewController(t))
	stream := glia.P2PStream{Stream: s}
	s.EXPECT().
		Protocol().
		Return(protocol.ID(proto)).
		AnyTimes()

	require.Equal(t, "Wt9hMLbqHmNuCsvqCW8AuKxUjwL", stream.ProcID())
	require.Equal(t, "echo", stream.MethodName())
}

func mkPeerID(t *testing.T) peer.ID {
	t.Helper()

	id, err := peer.Decode("12D3KooWFYcCMuKujeeDDPqnH6yHeVrnXPaCjfbYVmQ9fHxfRDtA")
	require.NoError(t, err)
	return id
}
