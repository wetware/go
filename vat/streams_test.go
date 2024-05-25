package vat_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/require"
	test_libp2p "github.com/wetware/go/test/libp2p"
	"github.com/wetware/go/vat"
)

func TestStreamHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := test_libp2p.NewMockConn(ctrl)
	c.EXPECT().RemotePeer().
		Return(peer.ID("test")).
		Times(1)

	s := test_libp2p.NewMockStream(ctrl)
	s.EXPECT().Conn().
		Return(c).
		Times(1)
	s.EXPECT().ID().
		Return("test").
		Times(1)
	s.EXPECT().Protocol().
		Return(protocol.ID("test")).
		Times(1)

	ch := make(chan network.Stream)
	handle := vat.NewStreamHandler(ctx, ch)
	go handle(s)

	_, err := vat.Listener{C: ch}.Accept(ctx)
	require.NoError(t, err)
}

// func TestXXX(t *testing.T) {
// 	t.Parallel()

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	h, err := libp2p.New(
// 		libp2p.NoTransports,
// 		libp2p.NoListenAddrs,
// 		libp2p.Transport(inproc.New()),
// 		libp2p.ListenAddrStrings("/inproc/~"))
// 	require.NoError(t, err)
// 	defer h.Close()

// 	r := wazero.NewRuntime(ctx)
// 	defer r.Close(ctx)

// 	wasi.MustInstantiate(ctx, r)

// 	sys, err := system.Builder{}.Instantiate(ctx, r)
// 	require.NoError(t, err)
// 	defer sys.Close(ctx)

// 	b, err := os.ReadFile("testdata/socket/main.wasm")
// 	require.NoError(t, err)

// 	cm, err := r.CompileModule(ctx, b)
// 	require.NoError(t, err)
// 	defer cm.Close(ctx)

// 	ch := make(chan []byte, 1)
// 	defer close(ch)

// 	mod, err := r.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
// 		WithName("test").
// 		WithStdin(sys.Stdin()).
// 		WithStdout(os.Stdout)) // support printf debugging in guest code
// 	require.NoError(t, err)
// 	defer mod.Close(ctx)

// 	proto := vat.ProtoFromModule(mod)
// 	handler := vat.HandlerConfig{
// 		Host:  h,
// 		Proto: proto,
// 	}.Build(ctx)
// 	defer handler.Release()

// 	h2, err := libp2p.New(
// 		libp2p.NoTransports,
// 		libp2p.NoListenAddrs,
// 		libp2p.Transport(inproc.New()),
// 		libp2p.ListenAddrStrings("/inproc/~"))
// 	require.NoError(t, err)
// 	defer h2.Close()

// 	err = h2.Connect(ctx, *host.InfoFromHost(h))
// 	require.NoError(t, err)
// 	go func() {
// 		s, err := h2.NewStream(ctx, h.ID(), proto)
// 		require.NoError(t, err)
// 		defer s.Close()

// 		<-ctx.Done()
// 	}()

// 	s, err := handler.Accept(ctx)
// 	require.NoError(t, err)
// 	require.NotNil(t, s)
// }
