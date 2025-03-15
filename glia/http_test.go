package glia_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/libp2p/go-libp2p"
	inproc "github.com/lthibault/go-libp2p-inproc-transport"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	glia "github.com/wetware/go/glia"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
)

// testEnv implements glia.Env for testing
type testEnv struct {
	logger *slog.Logger
	// Add other required fields based on the Env interface
}

func (e *testEnv) Log() *slog.Logger {
	return e.logger
}

// Add other required methods based on the Env interface

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	return &testEnv{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
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
	host, err := libp2p.New(libp2p.NoTransports, libp2p.NoListenAddrs)
	require.NoError(t, err)

	// Set up wazero runtime and load echo module
	r := wazero.NewRuntimeWithConfig(context.Background(), wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(context.Background())
	wasi_snapshot_preview1.MustInstantiate(context.Background(), r)

	bytecode, err := os.ReadFile("../examples/echo/main.wasm")
	require.NoError(t, err)

	cm, err := r.CompileModule(context.Background(), bytecode)
	require.NoError(t, err)
	defer cm.Close(context.Background())

	const expectedProc = "Wt9hMLbqHmNuCsvqCW8AuKxUjwL"
	pid, err := proc.ParsePID(expectedProc)
	require.NoError(t, err)

	p, err := proc.Command{
		PID:    pid,
		Stderr: os.Stderr,
	}.Instantiate(context.Background(), r, cm)
	require.NoError(t, err)
	defer p.Close(context.Background())

	h := &glia.HTTP{
		Env: &system.Env{
			Host: host,
		},
		Router: mockRouter{P: p},
	}
	h.Init()

	server := httptest.NewServer(h.Handler)
	defer server.Close()

	t.Run("status endpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/status")
		require.NoError(t, err, "Failed to get status")
		defer resp.Body.Close()

		require.Equal(t, http.StatusNoContent, resp.StatusCode)
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
		// Test POST request
		resp, err := http.Post(server.URL+"/ipfs/v0/host123/proc456/method789", "", nil)
		require.NoError(t, err, "Failed to post to glia endpoint")
		defer resp.Body.Close()

		require.NotEqual(t, http.StatusMethodNotAllowed, resp.StatusCode, "Expected endpoint to be found, got method not allowed")
	})

	t.Run("method not allowed", func(t *testing.T) {
		// Test wrong methods on endpoints
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
			server.URL+path.Join("/ipfs/v0", host.ID().String(), expectedProc, "echo"),
			"text/plain",
			strings.NewReader(message))
		require.NoError(t, err, "Failed to post to echo endpoint")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Failed to read response body")
		require.Equal(t, message, string(body), "Echo response should match input")
	})
}
