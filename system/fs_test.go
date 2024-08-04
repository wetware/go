package system_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/system"
)

func TestFS(t *testing.T) {
	t.Parallel()

	err := fstest.TestFS(system.FS{})
	require.NoError(t, err)
}
