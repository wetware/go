package system_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/ipfs/boxo/path"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/system"
)

const IPFS_ROOT = "/ipfs/QmcKcGYiGcfSQ1s3VX7SnzQwRZHsgAuuPXdnrRghjCrBMx/system/testdata/fs"

func TestFS(t *testing.T) {
	t.Parallel()

	root, err := path.NewPath(IPFS_ROOT)
	require.NoError(t, err)

	ipfs, err := rpc.NewLocalApi()
	require.NoError(t, err)

	fs := system.FS{Ctx: context.Background(), API: ipfs.Unixfs(), Root: root}
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
