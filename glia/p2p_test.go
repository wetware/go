package glia_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"os"
	"strings"
	"testing"

	"capnproto.org/go/capnp/v3"
	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/wetware/go/glia"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
	test_libp2p "github.com/wetware/go/test/libp2p"
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
	env := &system.Env{
		Host: h,
	}
	// buf := &bytes.Buffer{}

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

	// Instantiate glia P2P runtime and populate it with
	// the mock router.
	p2p := glia.P2P{
		Env:    env,
		Router: mockRouter{P: p},
	}

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
	err = p2p.ServeP2P(nopCloser{Writer: w}, &req)
	require.NoError(t, err)

	// Read the uvarint length prefix from the buffer
	n, err := binary.ReadUvarint(w)
	require.NoError(t, err,
		"failed to read uvarint length header from response buffer")
	b, err := io.ReadAll(io.LimitReader(w, int64(n)))
	require.NoError(t, err)

	m, err = capnp.Unmarshal(b)
	require.NoError(t, err)
	defer m.Release()

	res, err := glia.ReadRootResult(m)
	require.NoError(t, err)
	require.NotZero(t, res)
	// require.Equal(t, glia.Result_Status_ok, res.Status(), res.Status().String())
}

type mockRouter struct {
	P *proc.P
}

func (r mockRouter) GetProc(pid string) (glia.Proc, error) {
	return r.P, nil
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error {
	return nil
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

// type mockMethod struct {
// 	Body io.Reader
// }

// func (mm mockMethod) CallWithStack(context.Context, []uint64) error {
// 	_, err := io.Copy(io.Discard, mm.Body)
// 	return err
// }
