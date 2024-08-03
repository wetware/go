package guest

import (
	"errors"
	"io/fs"
)

var _ fs.FS = (*FS)(nil)

type FS struct{}

// Open opens the named file.
//
// When Open returns an error, it should be of type *PathError
// with the Op field set to "open", the Path field set to name,
// and the Err field describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// ValidPath(name), returning a *PathError with Err set to
// ErrInvalid or ErrNotExist.
func (FS) Open(name string) (fs.File, error) {
	return nil, errors.New("FS.Open::NOT IMPLEMENTED")
}
