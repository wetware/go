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

func TestFS(t *testing.T) {
	t.Parallel()

	root, err := path.NewPath("/ipfs/QmQuTsZYyFSVXD8r6yfWyJyJ5xhzV8wkqy9wWuTeoccDtW")
	require.NoError(t, err)

	ipfs, err := rpc.NewLocalApi()
	require.NoError(t, err)

	fs := system.FS{Ctx: context.Background(), API: ipfs.Unixfs(), Root: root}
	err = fstest.TestFS(fs,
		"testdata")
	require.NoError(t, err)
}

func TestIPFSNode(t *testing.T) {
	t.Parallel()

	root, err := path.NewPath("/ipfs/QmQuTsZYyFSVXD8r6yfWyJyJ5xhzV8wkqy9wWuTeoccDtW")
	require.NoError(t, err)

	ipfs, err := rpc.NewLocalApi()
	require.NoError(t, err)

	dir, err := ipfs.Unixfs().Ls(context.Background(), root)
	require.NoError(t, err)

	var names []string
	for entry := range dir {
		names = append(names, entry.Name)
	}

	expect := []string{"testdata"}
	require.ElementsMatch(t, names, expect,
		"unexpected file path")
}
