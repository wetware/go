// //go:generate capnp compile -I $GOPATH/src/capnproto.org/go/capnp/std -ogo system.capnp

package system

import (
	"context"
	"log/slog"

	"github.com/blang/semver/v4"
	"github.com/hashicorp/go-memdb"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
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

type Env struct {
	IPFS iface.CoreAPI
	Host host.Host
}

func (env Env) Log() *slog.Logger {
	return slog.With("peer", env.Host.ID())
}

func (env Env) HandlePeerFound(info peer.AddrInfo) {
	// TODO:  do we want to move this to boot/mdns.go?   Currently, this
	// callback is used exclusively by the MDNS discovery system, but it
	// can be used by other discovery systems in principle.

	pstore := env.Host.Peerstore()
	pstore.AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	env.Log().Info("peer discovered", "found", info.ID)
}
