package glia_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/glia"
)

func TestFrameIO(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	frame := []byte("test data")
	err := glia.WriteFrame(buf, frame)
	require.NoError(t, err)
	require.Equal(t, len(frame)+1, buf.Len())

	got, err := glia.ReadFrame(bufio.NewReader(buf))
	require.NoError(t, err)
	require.Equal(t, frame, got)
}
