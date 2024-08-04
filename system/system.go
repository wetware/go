package system

import (
	"io"
	"os"
)

type Streams struct {
	Reader    io.Reader
	Writer    io.WriteCloser
	ErrWriter io.WriteCloser
}

func (std Streams) Stdin() io.Reader {
	if std.Reader == nil {
		return os.Stdin
	}

	return std.Reader
}

func (std Streams) Stdout() io.WriteCloser {
	if std.Writer == nil {
		return os.Stdout
	}

	return std.Writer
}

func (std Streams) Stderr() io.WriteCloser {
	if std.ErrWriter == nil {
		return os.Stdout
	}

	return std.ErrWriter
}
