package ww_test

// import (
// 	"bytes"
// 	context "context"
// 	"io"
// 	"testing"

// 	"github.com/ipfs/boxo/path"
// 	"github.com/ipfs/kubo/client/rpc"
// 	"github.com/stretchr/testify/require"
// 	"github.com/tetratelabs/wazero/sys"
// 	ww "github.com/wetware/go"
// 	"github.com/wetware/go/system"
// )

// func TestService(t *testing.T) {
// 	t.Parallel()

// 	root, err := path.NewPath("/ipfs/QmRecDLNaESeNY3oUFYZKK9ftdANBB8kuLaMdAXMD43yon")
// 	require.NoError(t, err)

// 	rbuf := strings.NewReader("stdin test\n")
// 	wbuf := new(bytes.Buffer)
// 	ebuf := new(bytes.Buffer)

// 	cluster := ww.Config{
// 		NS:     root.String(),
// 		UnixFS: system.IPFS{
// 			Ctx: context.TODO(),
// 			Root: root,
// 			Unix: /*  TODO:  mock  */,
// 		},
// 		Stdio: struct {
// 			Reader io.Reader
// 			Writer io.WriteCloser
// 			Error  io.WriteCloser
// 		}{Reader: rbuf, Writer: wbuf, Error: ebuf},
// 	}.Build()

// 	err = cluster.Serve(context.Background())
// 	status := err.(*sys.ExitError).ExitCode()
// 	require.Zero(t, status)

// 	// Check that main.wasm wrote what we expect.
// 	require.Equal(t, "stdout test\n", wbuf.String())
// 	require.Equal(t, "stderr test\n", ebuf.String())
// }
