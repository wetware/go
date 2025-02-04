package glia_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"strings"
	"testing"

	"capnproto.org/go/capnp/v3"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/glia"
	"github.com/wetware/go/system"
	test_libp2p "github.com/wetware/go/test/libp2p"
)

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
	buf := &bytes.Buffer{}

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
		Method(gomock.Any()).          // context.Context
		Return(mockMethod{Body: buf}). // error
		After(reserve).
		Times(1)
	release := p.EXPECT().
		Release().
		After(method).
		Times(1)
	p.EXPECT().
		OutBuffer().
		Return(bytes.NewReader(nil)).
		After(release).
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

	// Read the uvarint length prefix from the buffer
	n, err := binary.ReadUvarint(w)
	require.NoError(t, err)

	b, err := io.ReadAll(io.LimitReader(w, int64(n)))
	require.NoError(t, err)

	m, err = capnp.Unmarshal(b)
	require.NoError(t, err)
	defer m.Release()

	res, err := glia.ReadRootResult(m)
	require.NoError(t, err)
	require.Equal(t, glia.Result_Status_ok, res.Status(), res.Status().String())
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
