package system

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/auth"
)

func TestReadPipe(t *testing.T) {
	t.Parallel()

	buf := newPipeBuffer(bytes.NewBufferString("test"))
	pipe := auth.ReadPipe_ServerToClient(buf)
	pr := pipeReader{ReadPipe: pipe}

	b, err := io.ReadAll(pr)
	require.NoError(t, err)
	require.Equal(t, "test", string(b))
}

func TestWritePipe(t *testing.T) {
	t.Parallel()

	buf := newPipeBuffer(new(bytes.Buffer))
	pipe := auth.WritePipe_ServerToClient(buf)
	pw := pipeWriter{WritePipe: pipe}

	n, err := io.Copy(pw, strings.NewReader("test"))
	require.NoError(t, err)
	require.Equal(t, len("test"), int(n))
	require.Equal(t, "test", buf.Buffer.String())
}

type pipeBuffer struct {
	Buffer *bytes.Buffer
}

func newPipeBuffer(b *bytes.Buffer) pipeBuffer {
	return pipeBuffer{Buffer: b}
}

func (p pipeBuffer) Read(ctx context.Context, read auth.ReadPipe_read) error {
	size := read.Args().Size()
	res, err := read.AllocResults()
	if err != nil {
		return err
	}

	r := io.LimitReader(p.Buffer, int64(size))
	b, err := io.ReadAll(r)
	if err == io.EOF {
		res.SetEof(true)
		err = nil
	} else if err == nil {
		err = res.SetData(b)
	}

	return err
}

func (p pipeBuffer) Write(ctx context.Context, write auth.WritePipe_write) error {
	data, err := write.Args().Data()
	if err != nil {
		return err
	}

	res, err := write.AllocResults()
	if err != nil {
		return err
	}

	n, err := io.Copy(p.Buffer, bytes.NewReader(data))
	if err == nil {
		res.SetN(uint32(n))
	}

	return err
}
