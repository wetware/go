package system

import (
	"context"
	"io/fs"
	"strings"

	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/pkg/errors"
)

type ReleaseFunc func()

var _ fs.FS = (*FSConfig)(nil)

type FSConfig struct {
	Ctx  context.Context
	Host host.Host
	IPFS iface.CoreAPI
}

func (fsc FSConfig) Open(name string) (fs.File, error) {
	if strings.HasPrefix(name, "/p2p/") {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  errors.New("TODO"),
		}

		// TODO:  uncomment

		// return PeerFS{
		// 	Ctx:  fsc.Ctx,
		// 	Host: fsc.Host,
		// }.Open(name)
	}

	return IPFS{
		Ctx:  fsc.Ctx,
		Unix: fsc.IPFS.Unixfs(),
	}.Open(name)

}
