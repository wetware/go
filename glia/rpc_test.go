package glia_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"capnproto.org/go/capnp/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/glia"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
	test_libp2p "github.com/wetware/go/test/libp2p"
)

func TestReadRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	call := newCallData()

	s := test_libp2p.NewMockStream(ctrl)
	s.EXPECT().
		Protocol().
		Return(system.Proto.Unwrap() + "/test-pid").
		Times(1)
	s.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(p []byte) (n int, err error) {
			return call.Read(p)
		}).
		MinTimes(1)

	req, err := glia.ReadRequest(context.TODO(), s)
	require.NoError(t, err)
	require.NotNil(t, req)
	require.Equal(t, "test-pid", req.PID)
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

	rpc := glia.RPC{
		Env:    env,
		Router: r,
	}

	calldata := newCallData()
	resdata := new(bytes.Buffer)

	s := test_libp2p.NewMockStream(ctrl)
	s.EXPECT().
		Protocol().
		Return(system.Proto.Unwrap() + "/test-pid").
		Times(1)
	s.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(p []byte) (n int, err error) {
			return calldata.Read(p)
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
	require.Equal(t, glia.Status_ok, res.Status())
}

func newCallData() *bytes.Buffer {
	m, s := capnp.NewSingleSegmentMessage(nil)

	cd, err := glia.NewRootCallData(s)
	if err != nil {
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
