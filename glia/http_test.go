package glia_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	glia "github.com/wetware/go/glia"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
)

func TestHTTP(t *testing.T) {
	t.Parallel()

	h := new(glia.HTTP)
	h.Init()
	server := httptest.NewServer(h.DefaultRouter())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedPeer := "12D3KooWPTR9RGhkm5D5XsJCMh2WGofMfTWcN4F79ofaScWGfEDw"
	expectedProc := "myProc"
	expectedMethod := "echo"
	// expectedStack := []uint64{1, 2, 3}
	expectedStackStr := "1,2,3"

	r := wazero.NewRuntimeWithConfig(context.TODO(), wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(context.TODO())
	wasi_snapshot_preview1.MustInstantiate(context.TODO(), r)

	bytecode, err := os.ReadFile("../examples/echo/main.wasm")
	require.NoError(t, err)

	cm, err := r.CompileModule(context.TODO(), bytecode)
	require.NoError(t, err)
	defer cm.Close(context.TODO())

	p, err := proc.Command{
		PID: proc.NewPID(),
		// Args: ,
		// Env: ,
		Stderr: os.Stderr,
		// FS: ,
	}.Instantiate(context.TODO(), r, cm)
	require.NoError(t, err)
	defer p.Close(context.TODO())

	mockRouter := NewMockRouter(ctrl)
	mockRouter.EXPECT().
		GetProc(expectedProc).
		Return(p, nil).
		Times(1)
	h.P2P.Router = mockRouter

	client := &http.Client{}
	url := server.URL + path.Join("/", system.Proto.String(), expectedPeer, expectedProc, expectedMethod)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "text/plain")

	restParams := req.URL.Query()
	restParams.Add("stack", expectedStackStr)
	req.URL.RawQuery = restParams.Encode()

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(res.Body)
		t.Fatalf("HTTP request failed with status: %d: %s", res.StatusCode, string(errMsg))
	}
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
