package system

import (
	"context"
	"io/fs"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/pkg/errors"
)

var _ fs.FS = (*FS)(nil)

// An FS provides access to a hierarchical file system.
//
// The FS interface is the minimum implementation required of the file system.
// A file system may implement additional interfaces,
// such as [ReadFileFS], to provide additional or optimized functionality.
//
// [testing/fstest.TestFS] may be used to test implementations of an FS for
// correctness.
type FS struct {
	API  iface.UnixfsAPI
	Path path.Path
}

// Open opens the named file.
//
// When Open returns an error, it should be of type *PathError
// with the Op field set to "open", the Path field set to name,
// and the Err field describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// fs.ValidPath(name), returning a *fs.PathError with Err set to
// fs.ErrInvalid or fs.ErrNotExist.
func (f FS) Open(name string) (fs.File, error) {
	p, err := path.Join(f.Path, name)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "path.Join",
			Path: name,
			Err:  err,
		}
	}

	if !fs.ValidPath(p.String()) {
		return nil, &fs.PathError{
			Op:   "fs.ValidPath",
			Path: name,
			Err:  errors.New("invalid path"),
		}
	}

	node, err := f.API.Get(context.TODO(), p)
	if err != nil {
		return nil, err
	}

	switch n := node.(type) {
	case files.File:
		return fileNode{File: n}, nil

	case files.Directory:
		defer n.Close()
		return nil, &fs.PathError{
			Op:   "FS.Open",
			Path: name,
			Err:  errors.New("node is a directory"),
		}

	default:
		panic(n) // unhandled type
	}
}

// fileNode provides access to a single file. The fs.File interface is the minimum
// implementation required of the file. Directory files should also implement [ReadDirFile].
// A file may implement io.ReaderAt or io.Seeker as optimizations.

type fileNode struct {
	files.File
}

func (n fileNode) Stat() (fs.FileInfo, error) {
	return nil, errors.New("fileNode.Stat::NOT IMPLEMENTED")
}
