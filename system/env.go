package system

import (
	"context"
	"io"
	"log/slog"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/pkg/errors"
)

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

func (env Env) Load(ctx context.Context, p string) ([]byte, error) {
	path, err := path.NewPath(p)
	if err != nil {
		return nil, err
	}

	node, err := env.IPFS.Unixfs().Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer node.Close()

	// TODO: improve
	switch n := node.(type) {
	case files.File:
		return io.ReadAll(n)
	case files.Directory:
		entries := n.Entries()
		for entries.Next() {
			if entries.Name() == "main.wasm" {
				return io.ReadAll(entries.Node().(io.Reader))
			}
		}
	}

	return nil, errors.New("not found")
}

func (env Env) NewUnixFS(ctx context.Context) UnixFS {
	return UnixFS{
		Ctx:  ctx,
		Unix: env.IPFS.Unixfs(),
	}
}
