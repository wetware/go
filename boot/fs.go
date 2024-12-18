package boot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/libp2p/go-libp2p/core/discovery"
)

const DefaultTimeout = time.Second * 10

var _ fs.FS = (*FS)(nil)

type FS struct {
	Discoverer discovery.Discoverer
	Timeout    time.Duration
}

func (boot FS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	if boot.Timeout <= 0 {
		boot.Timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), boot.Timeout)
	peers, err := boot.Discoverer.FindPeers(ctx, name)
	if err != nil {
		defer cancel()
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fmt.Errorf("find peers: %w", err),
		}
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		defer cancel()

		enc := json.NewEncoder(pw)
		for info := range peers {
			if err := enc.Encode(info); err != nil {
				slog.ErrorContext(ctx, "failed to encode peer addr info",
					"reason", err)
			}
		}
	}()

	f := &File{
		Path:       name,
		ReadCloser: pr,
	}
	return f, nil
}

type File struct {
	Path string
	io.ReadCloser
}

func (f File) Stat() (fs.FileInfo, error) {
	return f, nil
}

// fs.FileInfo
////

// base name of the file
func (f File) Name() string {
	return filepath.Base(f.Path)
}

// length in bytes for regular files; system-dependent for others
func (f File) Size() int64 {
	return -1
}

// file mode bits
func (f File) Mode() fs.FileMode {
	return 0
}

// modification time
func (f File) ModTime() time.Time {
	return time.Time{}
}

// abbreviation for Mode().IsDir()
func (f File) IsDir() bool {
	return false
}

// underlying data source (can return nil)
func (f File) Sys() any {
	return f.ReadCloser
}
