package system

import (
	"github.com/blang/semver/v4"
	"github.com/hashicorp/go-memdb"
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
