package glia_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	inproc "github.com/lthibault/go-libp2p-inproc-transport"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	glia "github.com/wetware/go/glia"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
)

type testFixture struct {
	t      *testing.T
	ctx    context.Context
	rt     wazero.Runtime
	module wazero.CompiledModule
	host   host.Host
	server *httptest.Server
}

func newTestFixture(t *testing.T) *testFixture {
	t.Helper()

	// Create host with inproc transport
	h, err := libp2p.New(
		libp2p.NoTransports,
		libp2p.NoListenAddrs,
		libp2p.Transport(inproc.New()),
		libp2p.ListenAddrStrings("/inproc/~"))
	require.NoError(t, err)

	// Set up wazero runtime
	ctx := context.Background()
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	// Load echo module
	bytecode, err := os.ReadFile("../examples/echo/main.wasm")
	require.NoError(t, err)

	module, err := rt.CompileModule(ctx, bytecode)
	require.NoError(t, err)

	return &testFixture{
		t:      t,
		ctx:    ctx,
		rt:     rt,
		module: module,
		host:   h,
	}
}

func (f *testFixture) Close() {
	if f.server != nil {
		f.server.Close()
	}
	if f.module != nil {
		f.module.Close(f.ctx)
	}
	if f.rt != nil {
		f.rt.Close(f.ctx)
	}
}

func (f *testFixture) NewProcess() (*proc.P, string) {
	f.t.Helper()
	pid := proc.NewPID()
	p, err := proc.Command{
		PID:    pid,
		Stderr: os.Stderr,
	}.Instantiate(f.ctx, f.rt, f.module)
	require.NoError(f.t, err)
	return p, pid.String()
}

func (f *testFixture) NewHTTPServer(p *proc.P) *httptest.Server {
	f.t.Helper()
	h := &glia.HTTP{
		Env: &system.Env{
			Host: f.host,
		},
		Router: mockRouter{P: p},
		Root:   p.String(),
	}
	h.Init()
	f.server = httptest.NewServer(h.Handler)
	return f.server
}

func (f *testFixture) NewPeerWithProcess() (host.Host, *proc.P, string) {
	f.t.Helper()
	peer, err := libp2p.New(
		libp2p.NoTransports,
		libp2p.NoListenAddrs,
		libp2p.Transport(inproc.New()),
		libp2p.ListenAddrStrings("/inproc/~"))
	require.NoError(f.t, err)

	p, pid := f.NewProcess()
	return peer, p, pid
}

func (f *testFixture) ConnectToPeer(h host.Host) {
	f.t.Helper()
	info := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	err := f.host.Connect(f.ctx, info)
	require.NoError(f.t, err)
}

func TestHTTP(t *testing.T) {
	t.Parallel()

	const (
		expectedProc   = "Wt9hMLbqHmNuCsvqCW8AuKxUjwL"
		expectedMethod = "echo"
		// expectedStack = []uint64{1, 2, 3}
		// expectedStackStr = "1,2,3"
	)

	h, err := libp2p.New(libp2p.NoTransports,
		libp2p.NoListenAddrs,
		libp2p.Transport(inproc.New()),
		libp2p.ListenAddrStrings("/inproc/~"))
	require.NoError(t, err)
	expectedPeer := h.ID().String()

	r := wazero.NewRuntimeWithConfig(context.TODO(), wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(context.TODO())
	wasi_snapshot_preview1.MustInstantiate(context.TODO(), r)

	bytecode, err := os.ReadFile("../examples/echo/main.wasm")
	require.NoError(t, err)

	cm, err := r.CompileModule(context.TODO(), bytecode)
	require.NoError(t, err)
	defer cm.Close(context.TODO())

	pid, err := proc.ParsePID(expectedProc)
	require.NoError(t, err)
	p, err := proc.Command{
		PID: pid,
		// Args: ,
		// Env: ,
		Stderr: os.Stderr,
		// FS: ,
	}.Instantiate(context.TODO(), r, cm)
	require.NoError(t, err)
	defer p.Close(context.TODO())

	g := &glia.HTTP{
		Env: &system.Env{
			Host: h,
			// IPFS: ,
		},
		Router: mockRouter{P: p},
	}
	g.Init()
	server := httptest.NewServer(g.DefaultRouter())

	client := &http.Client{}
	url := server.URL + path.Join("/", system.Proto.String(), expectedPeer, expectedProc, expectedMethod)
	body := bytes.NewBufferString("Hello, Wetware!")
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "text/plain")

	// restParams := req.URL.Query()
	// restParams.Add("stack", expectedStackStr)
	// req.URL.RawQuery = restParams.Encode()

	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	got, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "Hello, Wetware!", string(got))
}

func TestHTTPStream(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test/proc/method", strings.NewReader("test body"))

	stream := glia.HTTPStream{
		ResponseWriter: w,
		Request:        r,
	}

	t.Run("ProcID", func(t *testing.T) {
		r.SetPathValue("proc", "test-proc")
		require.Equal(t, "test-proc", stream.ProcID())
	})

	t.Run("MethodName", func(t *testing.T) {
		r.SetPathValue("method", "test-method")
		require.Equal(t, "test-method", stream.MethodName())
	})

	t.Run("Read", func(t *testing.T) {
		buf := make([]byte, 100)
		n, err := stream.Read(buf)
		require.NoError(t, err)
		require.Equal(t, "test body", string(buf[:n]))
	})

	t.Run("Write", func(t *testing.T) {
		body := "test response"
		n, err := stream.Write([]byte(body))
		require.NoError(t, err)
		require.Equal(t, len(body), n)
		require.Equal(t, body, w.Body.String())
	})

	t.Run("Close", func(t *testing.T) {
		require.NoError(t, stream.Close())
	})
}

func TestHTTPIntegration(t *testing.T) {
	f := newTestFixture(t)
	defer f.Close()

	// Create local process
	p, pid := f.NewProcess()
	defer p.Close(f.ctx)

	// Set up HTTP server
	server := f.NewHTTPServer(p)
	defer server.Close()

	t.Run("status endpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/status")
		require.NoError(t, err, "Failed to get status")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "text/plain", resp.Header.Get("Content-Type"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read response body")

		// Verify the response contains the API path with root PID
		expectedPath := path.Join(system.Proto.String(), f.host.ID().String(), pid)
		require.Equal(t, expectedPath, string(body), "Status should return API path with root PID")
	})

	t.Run("info endpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/info")
		require.NoError(t, err, "Failed to get info")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var info peer.AddrInfo
		err = json.NewDecoder(resp.Body).Decode(&info)
		require.NoError(t, err, "Failed to decode JSON response")
		require.Equal(t, f.host.ID(), info.ID)
		require.NotEmpty(t, info.Addrs, "Expected non-empty addresses")
	})

	t.Run("root endpoint", func(t *testing.T) {
		// Test root endpoint
		resp, err := http.Get(server.URL + "/root")
		require.NoError(t, err, "Failed to get root")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read response body")

		// Verify the response is a valid proc.ID
		pid, err := proc.ParsePID(string(body))
		require.NoError(t, err, "Response should be a valid proc.ID")
		require.Len(t, pid[:], 20, "proc.ID should be 20 bytes")

		// Test deterministic behavior - should yield same PID
		resp2, err := http.Get(server.URL + "/root")
		require.NoError(t, err, "Failed to get root second time")
		defer resp2.Body.Close()

		body2, err := io.ReadAll(resp2.Body)
		require.NoError(t, err, "Failed to read second response body")

		pid2, err := proc.ParsePID(string(body2))
		require.NoError(t, err, "Second response should be a valid proc.ID")
		require.Equal(t, pid, pid2, "Root endpoint should yield same PID")

		// Test with POST method (should fail)
		resp, err = http.Post(server.URL+"/root", "", nil)
		require.NoError(t, err, "Failed to post to root endpoint")
		defer resp.Body.Close()
		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("version endpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/version")
		require.NoError(t, err, "Failed to get version")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read response body")
		require.NotEmpty(t, body, "Expected non-empty version response")
	})

	t.Run("glia endpoint", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/ipfs/v0/host123/proc456/method789", "", nil)
		require.NoError(t, err, "Failed to post to glia endpoint")
		defer resp.Body.Close()

		require.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("method not allowed", func(t *testing.T) {
		endpoints := []string{"/status", "/version"}
		for _, endpoint := range endpoints {
			resp, err := http.Post(server.URL+endpoint, "", nil)
			require.NoError(t, err, "Failed to post to %s", endpoint)
			defer resp.Body.Close()

			require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode,
				"Expected status 405 for POST to %s", endpoint)
		}
	})

	t.Run("echo endpoint", func(t *testing.T) {
		const message = "Hello, Wetware!"
		resp, err := http.Post(
			server.URL+path.Join("/", system.Proto.String(), f.host.ID().String(), pid, "echo"),
			"text/plain",
			strings.NewReader(message))
		require.NoError(t, err, "Failed to post to echo endpoint")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read response body")
		require.Equal(t, message, string(body), "Echo response should match input")
	})

	t.Run("remote echo endpoint", func(t *testing.T) {
		// Create peer B with its own process
		peerB, pB, pidB := f.NewPeerWithProcess()
		defer pB.Close(f.ctx)

		// Set up P2P handlers on both hosts
		p2pA := &glia.P2P{
			Env: &system.Env{
				Host: f.host,
			},
			Router: mockRouter{P: p},
		}
		go p2pA.Serve(f.ctx)

		p2pB := &glia.P2P{
			Env: &system.Env{
				Host: peerB,
			},
			Router: mockRouter{P: pB},
		}
		go p2pB.Serve(f.ctx)

		// Create HTTP handler for host A that will route to host B
		serverA := f.NewHTTPServer(pB)
		defer serverA.Close()

		// Connect host A to host B
		f.ConnectToPeer(peerB)

		// Send request to host A that should be routed to process on host B
		const message = "Hello from remote ww host!"
		resp, err := http.Post(
			serverA.URL+path.Join("/", system.Proto.String(), peerB.ID().String(), pidB, "echo"),
			"text/plain",
			strings.NewReader(message))
		require.NoError(t, err, "failed to post to remote echo endpoint")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "failed to read response body")
		require.Equal(t, message, string(body), "remote echo response should match input")
	})
}
