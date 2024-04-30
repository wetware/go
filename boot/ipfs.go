package boot

import (
	"context"
	"log/slog"

	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
)

type IPFS struct {
	Host host.Host
	API  iface.CoreAPI
	CID  cid.Cid
}

func (i IPFS) Serve(ctx context.Context) error {
	e, err := i.Host.EventBus().Emitter(new(EvtPeerFound))
	if err != nil {
		return err
	}
	defer e.Close()

	n, err := i.API.ResolveNode(ctx, path.FromCid(i.CID))
	if err != nil {
		return err
	}

	slog.WarnContext(ctx, string(n.RawData()),
		"cid", n.Cid())

	// TODO:  use IPLD node.
	// ...

	<-ctx.Done()
	return ctx.Err()
}
