package system

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/hashicorp/go-memdb"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/wetware/go/proc"
)

var _ fs.FS = (*Anchor)(nil)

type Anchor struct {
	Ctx     context.Context
	IPFS    iface.CoreAPI
	Host    host.Host
	Routing *memdb.MemDB
}

// The public host is routed over IPFS.
func (root Anchor) Public() *routedhost.RoutedHost {
	public := root.IPFS.Routing()
	return routedhost.Wrap(root.Host, public)
}

// fs.FS methods
////

func (root Anchor) Open(name string) (f fs.File, err error) {
	switch {
	case name == ".":
		return root, nil

	case strings.HasPrefix(name, "p2p/"):
		return HostNode{
			Ctx:     root.Ctx,
			Host:    root.Public(),
			Routing: root.Routing,
		}.Open(name)

	case strings.HasPrefix(name, "ipfs/"):
		return UnixFS{
			Ctx:  root.Ctx,
			Unix: root.IPFS.Unixfs(),
		}.Open(name)

	default:
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrNotExist,
		}
	}
}

// fs.File methods
////

func (root Anchor) Close() error {
	return nil
}

func (root Anchor) Stat() (fs.FileInfo, error) {
	return root, nil
}

func (root Anchor) Read(b []byte) (int, error) {
	return 0, io.EOF
}

// fs.FileInfo methods
////

// base name of the file
func (root Anchor) Name() string {
	return "/p2p/" + root.Host.ID().String()
}

// length in bytes for regular files; system-dependent for others
func (root Anchor) Size() int64 {
	return 0
}

// file mode bits
func (root Anchor) Mode() fs.FileMode {
	return fs.ModeDir
}

// modification time
func (root Anchor) ModTime() time.Time {
	return time.Time{}
}

// abbreviation for Mode().IsDir()
func (root Anchor) IsDir() bool {
	return true
}

// underlying data source (can return nil)
func (root Anchor) Sys() any {
	return root.Host
}

var _ fs.File = (*StreamNode)(nil)

type StreamNode struct {
	network.Stream
}

// fs.File methods
////

func (s StreamNode) Stat() (fs.FileInfo, error) {
	return s, nil
}

// fs.FileInfo methods
////

// base name of the file
func (s StreamNode) Name() string {
	return s.Stream.ID()
}

// length in bytes for regular files; system-dependent for others
func (s StreamNode) Size() int64 {
	return 0
}

// file mode bits
func (s StreamNode) Mode() fs.FileMode {
	return 0
}

// modification time
func (s StreamNode) ModTime() time.Time {
	return time.Time{}
}

// abbreviation for Mode().IsDir()
func (s StreamNode) IsDir() bool {
	return false
}

// underlying data source (can return nil)
func (s StreamNode) Sys() any {
	return s.Stream
}

var _ fs.File = (*StreamNode)(nil)

type ProcNode struct {
	Ctx context.Context
	P   *proc.P
	Buf *bytes.Buffer
}

// fs.File methods
////

func (p ProcNode) Stat() (fs.FileInfo, error) {
	return p, nil
}

func (p ProcNode) Read(b []byte) (int, error) {
	return p.Buf.Read(b)
}

func (p ProcNode) Write(b []byte) (int, error) {
	return p.Buf.Write(b)
}

// Close flushes the data in the buffer, delivering it to the process,
// Close() DOES NOT close the underlying process.
func (p ProcNode) Close() error {
	// return p.P.Deliver(p.Ctx, p.Buf)
	panic("NOT IMPLEMENTED")
}

// fs.FileInfo methods
////

// base name of the file
func (p ProcNode) Name() string {
	return p.P.String()
}

// length in bytes for regular files; system-dependent for others
func (p ProcNode) Size() int64 {
	return 0
}

// file mode bits
func (p ProcNode) Mode() fs.FileMode {
	return 0
}

// modification time
func (p ProcNode) ModTime() time.Time {
	return time.Time{}
}

// abbreviation for Mode().IsDir()
func (p ProcNode) IsDir() bool {
	return false
}

// underlying data source (can return nil)
func (p ProcNode) Sys() any {
	return p.P
}
