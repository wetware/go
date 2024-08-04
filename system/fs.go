package system

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/pkg/errors"
)

var (
	_ fs.FS     = (*FS)(nil)
	_ fs.StatFS = (*FS)(nil)
)

// An FS provides access to a hierarchical file system.
//
// The FS interface is the minimum implementation required of the file system.
// A file system may implement additional interfaces,
// such as [ReadFileFS], to provide additional or optimized functionality.
//
// [testing/fstest.TestFS] may be used to test implementations of an FS for
// correctness.
type FS struct {
	Ctx  context.Context
	API  iface.UnixfsAPI
	Root path.Path
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
	if pathInvalid(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  errors.New("invalid path"),
		}
	}

	path, node, err := f.Resolve(name)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fmt.Errorf("resolve: %w", err),
		}
	}

	return &ipfsNode{
		Path: path,
		Node: node,
	}, nil
}

func (f FS) Resolve(name string) (path.Path, files.Node, error) {
	joined, err := path.Join(f.Root, clean(name))
	if err != nil {
		return nil, nil, err
	}

	node, err := f.API.Get(f.Ctx, joined)
	return joined, node, err
}

func pathInvalid(name string) bool {
	return !fs.ValidPath(name)
}

func clean(name string) string {
	name = filepath.Clean(name)
	// name = strings.Trim(name, "./")
	return name
}

// Stat returns a FileInfo describing the file.
// If there is an error, it should be of type *PathError.
func (f FS) Stat(name string) (fs.FileInfo, error) {
	path, node, err := f.Resolve(name)
	if err != nil {
		return nil, err
	}

	return &ipfsNode{
		Path: path,
		Node: node,
	}, nil
}

var (
	_ fs.FileInfo    = (*ipfsNode)(nil)
	_ fs.ReadDirFile = (*ipfsNode)(nil)
	_ fs.DirEntry    = (*ipfsNode)(nil)
)

// ipfsNode provides access to a single file. The fs.File interface is the minimum
// implementation required of the file. Directory files should also implement [ReadDirFile].
// A file may implement io.ReaderAt or io.Seeker as optimizations.
type ipfsNode struct {
	Path path.Path
	files.Node
}

// base name of the file
func (n ipfsNode) Name() string {
	segs := n.Path.Segments()
	return segs[len(segs)-1] // last segment is name
}

func (n *ipfsNode) Stat() (fs.FileInfo, error) {
	return n, nil
}

// length in bytes for regular files; system-dependent for others
func (n ipfsNode) Size() int64 {
	size, err := n.Node.Size()
	if err != nil {
		slog.Error("failed to obtain file size",
			"path", n.Path,
			"reason", err)
	}

	return size
}

// file mode bits
func (n ipfsNode) Mode() fs.FileMode {
	switch n.Node.(type) {
	case files.Directory:
		return fs.ModeDir
	default:
		return 0 // regular read-only file
	}
}

// modification time
func (n ipfsNode) ModTime() time.Time {
	return time.Time{} // zero-value time
}

// abbreviation for Mode().IsDir()
func (n ipfsNode) IsDir() bool {
	return n.Mode().IsDir()
}

// underlying data source (never returns nil)
func (n ipfsNode) Sys() any {
	return n.Node
}

func (n ipfsNode) Read(b []byte) (int, error) {
	switch node := n.Node.(type) {
	case io.Reader:
		return node.Read(b)
	default:
		return 0, errors.New("unreadable node")
	}
}

// ReadDir reads the contents of the directory and returns
// a slice of up to max DirEntry values in directory order.
// Subsequent calls on the same file will yield further DirEntry values.
//
// If max > 0, ReadDir returns at most max DirEntry structures.
// In this case, if ReadDir returns an empty slice, it will return
// a non-nil error explaining why.
// At the end of a directory, the error is io.EOF.
// (ReadDir must return io.EOF itself, not an error wrapping io.EOF.)
//
// If max <= 0, ReadDir returns all the DirEntry values from the directory
// in a single slice. In this case, if ReadDir succeeds (reads all the way
// to the end of the directory), it returns the slice and a nil error.
// If it encounters an error before the end of the directory,
// ReadDir returns the DirEntry list read until that point and a non-nil error.
func (n ipfsNode) ReadDir(max int) (entries []fs.DirEntry, err error) {
	root, ok := n.Node.(files.Directory)
	if !ok {
		return nil, errors.New("not a directory")
	}

	iter := root.Entries()
	for iter.Next() {
		name := iter.Name()
		node := iter.Node()

		// Callers will typically discard entries if they get a non-nill
		// error, so we make sure nodes are eventually closed.
		runtime.SetFinalizer(node, func(c io.Closer) {
			if err := c.Close(); err != nil {
				slog.Warn("unable to close node",
					"name", name,
					"reason", err)
			}
		})

		var subpath path.Path
		if subpath, err = path.Join(n.Path, name); err != nil {
			break
		}

		entries = append(entries, &ipfsNode{
			Path: subpath,
			Node: node})

		// got max items?
		if max--; max == 0 {
			return
		}
	}

	// If we get here, it's because the iterator stopped.  It either
	// failed or is exhausted.

	if err == nil && iter.Err() != nil {
		err = iter.Err() // iterator failed
	}

	return
}

// Info returns the FileInfo for the file or subdirectory described by the entry.
// The returned FileInfo may be from the time of the original directory read
// or from the time of the call to Info. If the file has been removed or renamed
// since the directory read, Info may return an error satisfying errors.Is(err, ErrNotExist).
// If the entry denotes a symbolic link, Info reports the information about the link itself,
// not the link's target.
func (n *ipfsNode) Info() (fs.FileInfo, error) {
	return n, nil
}

// Type returns the type bits for the entry.
// The type bits are a subset of the usual FileMode bits, those returned by the FileMode.Type method.
func (n ipfsNode) Type() fs.FileMode {
	if n.Mode().IsDir() {
		return fs.ModeDir
	}

	return 0
}
