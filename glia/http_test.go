package glia_test

import (
	"testing"
)

// import (
// 	"bytes"
// 	"context"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"path"
// 	"strings"
// 	"testing"

// 	capnp "capnproto.org/go/capnp/v3"
// 	"github.com/go-chi/render"
// 	"github.com/golang/mock/gomock"
// 	"github.com/libp2p/go-libp2p/core/host"
// 	"github.com/libp2p/go-libp2p/core/network"
// 	"github.com/libp2p/go-libp2p/core/peer"
// 	"github.com/libp2p/go-libp2p/core/protocol"
// 	"github.com/tj/assert"

// 	"github.com/stretchr/testify/require"
// 	"github.com/wetware/go/glia"
// 	"github.com/wetware/go/system"
// 	test_libp2p "github.com/wetware/go/test/libp2p"
// 	protoutils "github.com/wetware/go/util/proto"
// )

func TestHTTP(t *testing.T) {
	t.Parallel()

}

// func TestHTTP_DefaultRouter(t *testing.T) {
// 	t.Parallel()

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	h := test_libp2p.NewMockHost(ctrl)
// 	// TODO:  expect function calls to h

// 	server := &glia.HTTP{
// 		P2P: glia.P2P{
// 			Env: &system.Env{
// 				Host: h,
// 			},
// 		},
// 	}

// 	t.Run("DeliverMessage", func(t *testing.T) {
// 		body := strings.NewReader("test data")
// 		r := newPostRequest(body)
// 		w := httptest.NewRecorder()

// 		server.DefaultRouter().ServeHTTP(w, r)
// 		require.Equal(t, http.StatusOK, w.Code)

// 		got, err := io.ReadAll(r.Body)
// 		require.NoError(t, err)
// 		_ = got // TODO:  when we add response data, check it
// 	})

// 	t.Run("BadMethod", func(t *testing.T) {
// 		body := strings.NewReader("test data")
// 		r := newPostRequest(body)
// 		r.Method = http.MethodGet // invalid

// 		w := httptest.NewRecorder()

// 		server.DefaultRouter().ServeHTTP(w, r)
// 		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
// 	})
// }

// func TestBindHTTPRequest(t *testing.T) {
// 	t.Parallel()

// 	body := strings.NewReader("test-body")
// 	r := newPostRequest(body)
// 	var req glia.MessageRoutingRequest
// 	require.NoError(t, req.Bind(r), "failed to bind *http.Request to HTTPRequest")
// 	require.Equal(t, r, req.HTTP)
// 	require.NotNil(t, r.Body,
// 		"body should not be empty")

// 	err := req.Bind(r)
// 	require.NoError(t, err)
// 	require.Equal(t, r.URL, req.HTTP.URL)
// 	require.Equal(t, r.Body, req.HTTP.Body)
// }

// func TestBindAndDeliverMessage(t *testing.T) {
// 	t.Parallel()

// 	t.Run("RemotePeer", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		h := test_libp2p.NewMockHost(ctrl)

// 		deliver, release := newBindAndDeliverMessage(h)
// 		defer release()

// 		id, err := deliver.RemotePeer()
// 		got := id.String()
// 		want := "12D3KooWCvQ2ZqbDvYKWBoQbc1zqCYjZ2rmU5hTTnQPGhA86WFxh"
// 		require.NoError(t, err)
// 		require.Equal(t, want, got)
// 	})

// 	t.Run("Proto", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		h := test_libp2p.NewMockHost(ctrl)

// 		deliver, release := newBindAndDeliverMessage(h)
// 		defer release()

// 		proto := system.Proto.Unwrap() // /ww/<version>
// 		want := protoutils.Join(proto, "glia", "test-pid")
// 		got := deliver.Proto()
// 		require.Equal(t, want, got)
// 	})

// 	t.Run("Render", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		buf := new(bytes.Buffer)

// 		h := test_libp2p.NewMockHost(ctrl)
// 		h.EXPECT().
// 			NewStream(context.Background(), gomock.Any(), gomock.Any()).
// 			DoAndReturn(func(_ context.Context, id peer.ID, ps ...protocol.ID) (network.Stream, error) {
// 				require.Equal(t, "12D3KooWCvQ2ZqbDvYKWBoQbc1zqCYjZ2rmU5hTTnQPGhA86WFxh", id.String())
// 				return streamBuffer{Buf: buf}, nil
// 			}).
// 			Times(1)

// 		deliver, release := newBindAndDeliverMessage(h)
// 		defer release()

// 		r := deliver.Req.HTTP
// 		w := httptest.NewRecorder()
// 		err := render.Render(w, r, deliver)
// 		require.NoError(t, err)

// 		// Check that the stream contains a well-formed CallData message.
// 		////
// 		m, err := capnp.Unmarshal(buf.Bytes())
// 		require.NoError(t, err)
// 		defer m.Release()

// 		call, err := glia.ReadRootCallData(m)
// 		require.NoError(t, err)
// 		require.True(t, call.IsValid())
// 	})
// }

// func newTestURL() *url.URL {
// 	proto := path.Join(system.Proto.String(), "glia")
// 	u := &url.URL{
// 		Host: "localhost:2080",
// 		Path: path.Join(proto, "test-pid"),
// 	}

// 	q := u.Query()
// 	q.Set("method", "test-method")
// 	q.Set("host", "12D3KooWCvQ2ZqbDvYKWBoQbc1zqCYjZ2rmU5hTTnQPGhA86WFxh")
// 	u.RawQuery = q.Encode()

// 	return u
// }

// func newPostRequest(body io.Reader) *http.Request {
// 	return httptest.NewRequest("POST", newTestURL().String(), body)
// }

// func newBindAndDeliverMessage(h host.Host) (*glia.BindAndDeliverMessage, capnp.ReleaseFunc) {
// 	m, seg := capnp.NewSingleSegmentMessage(nil)
// 	// NOTE:  m is released via the capnp.ReleaseFunc return value.

// 	body := strings.NewReader("test-body")
// 	r := newPostRequest(body)

// 	var req glia.MessageRoutingRequest
// 	if err := req.Bind(r); err != nil {
// 		panic(err)
// 	}

// 	if err := glia.BindRequestToHeaders(seg, req); err != nil {
// 		panic(err)
// 	}

// 	return &glia.BindAndDeliverMessage{
// 		Host:    h,
// 		Message: m,
// 		Req:     req,
// 	}, m.Release
// }

// func TestNewBindAndDeliverMessage(t *testing.T) {
// 	assert.NotPanics(t, func() {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()

// 		h := test_libp2p.NewMockHost(ctrl)

// 		deliver, release := newBindAndDeliverMessage(h)
// 		defer release()
// 		assert.NotZero(t, *deliver) // deref; ensure concrete value is initialized
// 	})
// }

// type streamBuffer struct {
// 	Buf *bytes.Buffer
// 	network.Stream
// }

// func (sb streamBuffer) Write(b []byte) (int, error) {
// 	return sb.Buf.Write(b)
// }

// func (sb streamBuffer) Close() error {
// 	return nil
// }
