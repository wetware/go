package guest_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/guest"
)

func TestFS(t *testing.T) {
	t.Parallel()

	err := fstest.TestFS(guest.FS{})
	require.NoError(t, err)
}
