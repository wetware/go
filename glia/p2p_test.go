package glia_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"capnproto.org/go/capnp/v3"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"github.com/thejerf/suture/v4"
	"github.com/wetware/go/glia"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
	test_libp2p "github.com/wetware/go/test/libp2p"
)

func TestGliaRPC(t *testing.T) {
	t.Parallel()

	// Set up our mocking infrastructure.  We'll be mocking
	// three major interfaces:
	//  1. host.Host
	//  2. glia.Router
	//  3. glia.Proc
	////
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Initialize libp2p mocks.
	h := test_libp2p.NewMockHost(ctrl)
	env := &system.Env{
		Host: h,
	}

	// Initialize wetware mocks.  Note the initialization in
	// reverse-call-order.  This is because we are testing a
	// 'happens after' ordering (>) such that:
	//    release > method > reserve
	p := NewMockProc(ctrl)
	reserve := p.EXPECT().
		Reserve(gomock.Any(), gomock.Any()). // context.Context, io.Reader
		Return(nil).                         // error
		Times(1)
	method := p.EXPECT().
		Method(gomock.Any()). // context.Context
		Return(nil).          // error
		After(reserve).
		Times(1)
	p.EXPECT().
		Release().
		After(method).
		Times(1)

	r := NewMockRouter(ctrl)
	r.EXPECT().
		GetProc("test-proc").
		Return(p, nil).
		Times(1)

	// Instantiate glia P2P runtime and populate it with
	// the mock router.
	p2p := glia.P2P{
		Env:    env,
		Router: r,
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	m, seg := capnp.NewSingleSegmentMessage(nil)
	defer m.Release()

	hdr, err := glia.NewRootHeader(seg)
	require.NoError(t, err)

	id, err := peer.Decode("12D3KooWPTR9RGhkm5D5XsJCMh2WGofMfTWcN4F79ofaScWGfEDw")
	require.NoError(t, err)

	require.NoError(t, hdr.SetPeer([]byte(id)))
	require.NoError(t, hdr.SetProc("test-proc"))
	require.NoError(t, hdr.SetMethod("test-method"))
	// hdr.SetStack()

	body := "hello, Wetware!"

	req := glia.Request{
		Ctx:    ctx,
		Header: hdr,
	}
	req.Body.Reset(strings.NewReader(body))

	w := &bytes.Buffer{}
	err = p2p.ServeP2P(w, &req)
	require.NoError(t, err)
}

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

	// Initialize libp2p mocks.
	h := test_libp2p.NewMockHost(ctrl)
	env := &system.Env{
		Host: h,
	}

	// Initialize wetware mocks.  Note the initialization in
	// reverse-call-order.  This is because we are testing a
	// 'happens after' ordering (>) such that:
	//    release > method > reserve
	p := NewMockProc(ctrl)
	reserve := p.EXPECT().
		Reserve(gomock.Any(), gomock.Any()). // context.Context, io.Reader
		Return(nil).                         // error
		Times(1)
	method := p.EXPECT().
		Method(gomock.Any()). // context.Context
		Return(nil).          // error
		After(reserve).
		Times(1)
	p.EXPECT().
		Release().
		After(method).
		Times(1)

	r := NewMockRouter(ctrl)
	r.EXPECT().
		GetProc("test-proc").
		Return(p, nil).
		Times(1)

	// Instantiate glia P2P runtime and populate it with
	// the mock router.
	p2p := glia.P2P{
		Env:    env,
		Router: r,
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	m, seg := capnp.NewSingleSegmentMessage(nil)
	defer m.Release()

	hdr, err := glia.NewRootHeader(seg)
	require.NoError(t, err)

	id, err := peer.Decode("12D3KooWPTR9RGhkm5D5XsJCMh2WGofMfTWcN4F79ofaScWGfEDw")
	require.NoError(t, err)

	require.NoError(t, hdr.SetPeer([]byte(id)))
	require.NoError(t, hdr.SetProc("test-proc"))
	require.NoError(t, hdr.SetMethod("test-method"))
	// hdr.SetStack()

	suture.NewSimple("test-proc")

	body := "hello, Wetware!"

	req := glia.Request{
		Ctx:    ctx,
		Header: hdr,
	}
	req.Body.Reset(strings.NewReader(body))

	w := &bytes.Buffer{}
	err = p2p.ServeP2P(w, &req)
	require.NoError(t, err)

	m, err = capnp.Unmarshal(w.Bytes())
	require.NoError(t, err)
	defer m.Release()

	res, err := glia.ReadRootResult(m)
	require.NoError(t, err)
	require.NotZero(t, res)

	// TODO:  eventually, we should investigate the contents of
	// the result
}

func TestReadRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hdr := newTestHeader()

	s := test_libp2p.NewMockStream(ctrl)
	s.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(p []byte) (n int, err error) {
			return hdr.Read(p)
		}).
		MinTimes(1)

	req, err := glia.ReadRequest(context.TODO(), s)
	require.NoError(t, err)
	require.NotNil(t, req)

	pid, err := req.Header.Proc()
	require.NoError(t, err)
	require.Equal(t, "test-pid", pid)
}

func TestRPC_ServeStream(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := test_libp2p.NewMockHost(ctrl)
	env := &system.Env{
		Host: h,
	}

	var body io.Reader
	p := NewMockProc(ctrl)
	p.EXPECT().
		Reserve(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, r io.Reader) error {
			body = r
			return nil
		}).
		Times(1)

	call := p.EXPECT().
		Method("test").
		DoAndReturn(func(name string) proc.Method {
			if name != "test" {
				return nil
			}
			return &mockMethod{Body: body}
		}).
		Times(1)

	p.EXPECT().
		Release().
		After(call).
		Times(1)

	r := NewMockRouter(ctrl)
	r.EXPECT().
		GetProc("test-pid").
		Return(p, nil).
		Times(1)

	rpc := glia.P2P{
		Env:    env,
		Router: r,
	}

	hdr := newTestHeader()
	resdata := new(bytes.Buffer)

	s := test_libp2p.NewMockStream(ctrl)
	s.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(p []byte) (n int, err error) {
			return hdr.Read(p)
		}).
		MinTimes(1)
	write := s.EXPECT().
		Write(gomock.Any()).
		DoAndReturn(func(p []byte) (n int, err error) {
			return resdata.Write(p)
		}).
		MinTimes(1)
	s.EXPECT().
		CloseWrite().
		After(write).
		Times(1)
	err := rpc.ServeStream(ctx, s)
	require.NoError(t, err)

	m, err := capnp.Unmarshal(resdata.Bytes())
	require.NoError(t, err)
	defer m.Release()

	res, err := glia.ReadRootResult(m)
	require.NoError(t, err)
	require.NotZero(t, res)
	require.Equal(t, glia.Result_Status_ok, res.Status())
}

// newTestHeader returns a buffer containing a pre-populated
// header.  The caller is responsible for releasing m.
func newTestHeader() *bytes.Buffer {
	m, s := capnp.NewSingleSegmentMessage(nil)
	// m is released by caller

	cd, err := glia.NewRootHeader(s)
	if err != nil {
		panic(err)
	}

	if err := cd.SetProc("test-pid"); err != nil {
		panic(err)
	}

	if err := cd.SetMethod("test"); err != nil {
		panic(err)
	}

	stack, err := cd.NewStack(4)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 4; i++ {
		stack.Set(i, uint64(i))
	}

	var buf bytes.Buffer
	if err := glia.WriteMessage(&buf, m); err != nil {
		panic(err)
	}

	return &buf
}

type mockMethod struct {
	Body io.Reader
}

func (mm mockMethod) CallWithStack(context.Context, []uint64) error {
	_, err := io.Copy(io.Discard, mm.Body)
	return err
}
