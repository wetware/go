// //go:generate capnp compile -I $GOPATH/src/capnproto.org/go/capnp/std -ogo system.capnp

package system

import (
	"context"
	"io"

	"github.com/blang/semver/v4"
	"github.com/hashicorp/go-memdb"
	"github.com/tetratelabs/wazero/api"
	"github.com/wetware/go/proc"
	protoutils "github.com/wetware/go/util/proto"
)

const Version = "0.1.0"

var Proto = protoutils.VersionedID{
	ID:      "ww",
	Version: semver.MustParse(Version),
}

var Schema = memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		"proc": {
			Name: "proc",
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: proc.PIDIndexer{},
				},
				// "path": {
				// 	Name:    "path",
				// 	Unique:  true,
				// 	Indexer: PathIndexer{},
				// },
			},
		},
	},
}

type CloserFunc func(context.Context) error

func (close CloserFunc) Close(ctx context.Context) error {
	return close(ctx)
}

type Proc interface {
	Reserve(context.Context, io.ReadWriteCloser) error
	Release()

	api.Closer
	String() string
	Method(name string) proc.Method
}

type Router interface {
	GetProc(pid string) (Proc, error)
}
