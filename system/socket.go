package system

import (
	"context"
	"io"

	"github.com/wetware/go/auth"
)

var _ = auth.Session{Sock: (*Socket)(nil)}

type Socket struct{}

func (s Socket) Bind(ctx context.Context, pipe auth.WritePipe) io.WriteCloser {
	return pipeWriter{
		Context:   ctx,
		WritePipe: pipe,
	}
}

func (s Socket) Connect(ctx context.Context, pipe auth.ReadPipe) io.Reader {
	return pipeReader{
		Context:  ctx,
		ReadPipe: pipe,
	}
}

type pipeReader struct {
	Context context.Context
	auth.ReadPipe
}

func (r pipeReader) Read(p []byte) (n int, err error) {
	f, release := r.ReadPipe.Read(r.Context, func(read auth.ReadPipe_read_Params) error {
		read.SetSize(uint32(len(p)))
		return nil
	})
	defer release()

	var res auth.ReadPipe_read_Results
	if res, err = f.Struct(); err != nil {
		return
	}

	var b []byte
	if b, err = res.Data(); err == nil {
		n = copy(p, b)
	}

	if res.Eof() {
		err = io.EOF
	}

	return
}

type pipeWriter struct {
	Context context.Context
	auth.WritePipe
}

func (w pipeWriter) Write(p []byte) (n int, err error) {
	f, release := w.WritePipe.Write(w.Context, func(write auth.WritePipe_write_Params) error {
		return write.SetData(p)
	})
	defer release()

	var res auth.WritePipe_write_Results
	if res, err = f.Struct(); err == nil {
		n = int(res.N())
	}

	return
}

func (w pipeWriter) Close() error {
	return nil
}
