package guest_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/guest"
)

func TestFS(t *testing.T) {
	t.Parallel()

	_, err := guest.FS{}.Open("")
	require.EqualError(t, err, "FS.Open::NOT IMPLEMENTED")
}
