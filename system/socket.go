package system

import (
	"bytes"
	"context"
	"io"
	"log/slog"

	"capnproto.org/go/capnp/v3/exp/bufferpool"
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
		read.SetSize(int64(len(p)))
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

type socketReader struct{ io.Reader }

func NewReadPipe(r io.Reader) auth.ReadPipe {
	return auth.ReadPipe_ServerToClient(socketReader{Reader: r})
}

func (r socketReader) Read(ctx context.Context, read auth.ReadPipe_read) error {
	res, err := read.AllocResults()
	if err != nil {
		return err
	}

	size := int(read.Args().Size())
	buf := bufferpool.Default.Get(size)
	defer bufferpool.Default.Put(buf)

	n, err := r.Reader.Read(buf)
	if res.SetEof(err == io.EOF); res.Eof() {
		err = nil
	}

	if n > 0 {
		err = res.SetData(buf[:n])
	}

	return err
}

type socketWriter struct{ io.WriteCloser }

func NewWritePipe(wc io.WriteCloser) auth.WritePipe {
	return auth.WritePipe_ServerToClient(socketWriter{WriteCloser: wc})
}

func (w socketWriter) Shutdown() {
	if err := w.Close(); err != nil {
		slog.Error("failed to close writer",
			"reason", err)
	}
}

func (w socketWriter) Write(ctx context.Context, write auth.WritePipe_write) error {
	b, err := write.Args().Data()
	if err != nil {
		return err
	}

	res, err := write.AllocResults()
	if err != nil {
		return err
	}

	if n, err := io.Copy(w.WriteCloser, bytes.NewReader(b)); err == nil {
		res.SetN(n)
	}

	return err
}
