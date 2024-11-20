package system

import (
	"context"
	"io/fs"
	"strings"
	"time"

	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

var _ fs.FS = (*Anchor)(nil)

type Anchor struct {
	Ctx  context.Context
	IPFS iface.CoreAPI
	Host host.Host
}

func (root Anchor) Open(name string) (f fs.File, err error) {
	switch {
	case strings.HasPrefix(name, "/p2p/"):
		return HostNode{
			Ctx:  root.Ctx,
			Host: root.Host,
		}.Open(name)

	case strings.HasPrefix(name, "/ipfs/"):
		return UnixFS{
			Ctx:  root.Ctx,
			Unix: root.IPFS.Unixfs(),
		}.Open(name)

	default:
		return nil, fs.ErrNotExist
	}
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
