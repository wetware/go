package system_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/auth"
	"github.com/wetware/go/system"
)

func TestReadPipe(t *testing.T) {
	t.Parallel()

	pipe := system.NewReadPipe(strings.NewReader("test"))
	f, release := pipe.Read(context.TODO(), func(read auth.ReadPipe_read_Params) error {
		read.SetSize(int64(len("test")))
		return nil
	})
	defer release()

	res, err := f.Struct()
	require.NoError(t, err)
	data, err := res.Data()
	require.NoError(t, err)
	require.Equal(t, "test", string(data))
	require.False(t, res.Eof())

	f, release = pipe.Read(context.TODO(), func(read auth.ReadPipe_read_Params) error {
		read.SetSize(int64(len("test")))
		return nil
	})
	defer release()

	res, err = f.Struct()
	// require.EqualError(t, err, "auth.capnp:ReadPipe.read: EOF")
	// require.NoError(t, err)
	require.True(t, res.Eof(), "should report EOF")
}

func TestWritePipe(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	pipe := system.NewWritePipe(nopCloser{Writer: buf})
	f, release := pipe.Write(context.TODO(), func(write auth.WritePipe_write_Params) error {
		return write.SetData([]byte("test"))
	})
	defer release()

	res, err := f.Struct()
	require.NoError(t, err)
	require.Equal(t, int64(len("test")), res.N())
}

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }
