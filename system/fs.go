package system

import (
	"context"
	"errors"
	"io/fs"
	"os"

	iface "github.com/ipfs/kubo/core/coreiface"
)

var _ fs.FS = (*FS)(nil)

type FS struct {
	Ctx  context.Context
	IPFS iface.CoreAPI
}

func (fs FS) Open(name string) (fs.File, error) {
	return nil, &os.PathError{
		Op:   "open",
		Path: name,
		Err:  errors.New("fopen::NOT IMPLEMENTED"),
	}
}
