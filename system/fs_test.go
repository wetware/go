package system_test

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/ipfs/boxo/path"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/system"
)

const IPFS_ROOT = "/ipfs/QmRecDLNaESeNY3oUFYZKK9ftdANBB8kuLaMdAXMD43yon" // go/system/testdata/fs

func TestFS(t *testing.T) {
	t.Parallel()

	root, err := path.NewPath(IPFS_ROOT)
	require.NoError(t, err)

	ipfs, err := rpc.NewLocalApi()
	require.NoError(t, err)

	fs := system.FS{Ctx: context.Background(), Unix: ipfs.Unixfs(), Root: root}
	err = fstest.TestFS(fs,
		"main.go",
		"main.wasm",
		"testdata")
	require.NoError(t, err)
}

func TestIPFSNode(t *testing.T) {
	t.Parallel()

	root, err := path.NewPath(IPFS_ROOT)
	require.NoError(t, err)

	ipfs, err := rpc.NewLocalApi()
	require.NoError(t, err)

	dir, err := ipfs.Unixfs().Ls(context.Background(), root)
	require.NoError(t, err)

	var names []string
	for entry := range dir {
		names = append(names, entry.Name)
	}

	expect := []string{"testdata", "main.go", "main.wasm"}
	require.ElementsMatch(t, names, expect,
		"unexpected file path")
}

// TestSubFS ensures that the filesystem retunred by fs.Sub correctly
// handles the '.' path. The returned filesystem MUST ensure that '.'
// resolves to the root IPFS path.
func TestSubFS(t *testing.T) {
	t.Parallel()

	root, err := path.NewPath("/ipfs/QmSAyttKvYkSCBTghuMxAJaBZC3jD2XLRCQ5FB3CTrb9rE") // go/system/testdata
	require.NoError(t, err)

	ipfs, err := rpc.NewLocalApi()
	require.NoError(t, err)

	fs, err := fs.Sub(system.FS{Ctx: context.Background(), Unix: ipfs.Unixfs(), Root: root}, "fs")
	require.NoError(t, err)
	require.NotNil(t, fs)

	err = fstest.TestFS(fs,
		"main.go",
		"main.wasm",
		"testdata")
	require.NoError(t, err)
}
