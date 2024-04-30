package boot

import (
	"context"

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

	p := path.FromCid(i.CID)

	// _, err = i.API.ResolveNode(ctx, p)
	// if err != nil {
	// 	return err
	// }

	if err := i.API.Pin().Add(ctx, p); err != nil {
		return err
	}

	ps, err := i.API.Routing().FindProviders(ctx, p)
	if err != nil {
		return err
	}

	for info := range ps {
		EmitPeerFound{
			Emitter: e,
		}.HandlePeerFound(info)
	}

	<-ctx.Done()
	return ctx.Err()
}
