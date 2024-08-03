package guest

import (
	"context"
	"errors"
	"io/fs"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
)

var _ fs.FS = (*FS)(nil)

type FS struct {
	IPFS iface.CoreAPI
}

// Open opens the named file.
//
// When Open returns an error, it should be of type *PathError
// with the Op field set to "open", the Path field set to name,
// and the Err field describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// ValidPath(name), returning a *PathError with Err set to
// ErrInvalid or ErrNotExist.
func (fs FS) Open(name string) (fs.File, error) {
	p, err := path.NewPath(name)
	if err != nil {
		return nil, err
	}

	n, err := fs.IPFS.Unixfs().Get(context.TODO(), p)
	return fsNode{Node: n}, err
}

// fsNode provides access to a single file. The fs.File interface is the minimum
// implementation required of the file. Directory files should also implement [ReadDirFile].
// A file may implement io.ReaderAt or io.Seeker as optimizations.

type fsNode struct {
	files.Node
}

func (n fsNode) Stat() (fs.FileInfo, error) {
	return nil, errors.New("fsNode.Stat::NOT IMPLEMENTED")
}

func (n fsNode) Read([]byte) (int, error) {
	return 0, errors.New("fsNode.Read::NOT IMPLEMENTED")
}

func (n fsNode) Close() error {
	return n.Node.Close()
}
