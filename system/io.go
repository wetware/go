package system

import "context"

type pipeReader struct {
	Context context.Context
	ReadPipe
}

func (r pipeReader) Read(p []byte) (n int, err error) {
	f, release := r.ReadPipe.Read(r.Context, func(read ReadPipe_read_Params) error {
		read.SetSize(uint32(len(p)))
		return nil
	})
	defer release()

	res, err := f.Struct()
	if err != nil {
		return 0, err
	}

	var b []byte
	if b, err = res.Data(); err == nil {
		n = copy(p, b)
	}

	return
}

type pipeWriter struct {
	Context context.Context
	WritePipe
}

func (w pipeWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	err = w.WritePipe.Write(w.Context, func(write WritePipe_write_Params) error {
		return write.SetData(p)
	})

	return
}

func (w pipeWriter) Close() error {
	return nil
}
