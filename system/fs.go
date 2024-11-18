package system

import (
	"context"
	"io/fs"
	"strings"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/pkg/errors"
)

var _ fs.FS = (*FS)(nil)

type FS struct {
	Ctx  context.Context
	Host host.Host
	IPFS iface.CoreAPI
}

func (sys FS) Open(name string) (fs.File, error) {
	switch {
	case strings.HasPrefix(name, "/p2p/"):
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  errors.New("TODO"), // TODO(soon)
		}

	case strings.HasPrefix(name, "/ipfs/"):
		p, err := path.NewPath(name)
		if err != nil {
			return nil, err
		}

		n, err := sys.OpenUnix(sys.Ctx, p)
		if err != nil {
			return nil, &fs.PathError{
				Op:   "open",
				Path: name,
				Err:  err,
			}
		}

		return &UnixNode{
			Path: p,
			Node: n,
		}, nil
	}

	return nil, &fs.PathError{
		Op:   "open",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

func (sys FS) OpenUnix(ctx context.Context, p path.Path) (files.Node, error) {
	return sys.IPFS.Unixfs().Get(ctx, p)
}
