package boot

import (
	"context"

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
	// e, err := i.Host.EventBus().Emitter(new(EvtPeerFound))
	// if err != nil {
	// 	return err
	// }
	// defer e.Close()

	// p := path.FromCid(i.CID)

	// if err := i.API.Pin().Add(ctx, p, func(pas *options.PinAddSettings) error {
	// 	// pas.Name =
	// 	// pas.Recursive =
	// 	return nil
	// }); err != nil {
	// 	return err
	// }

	// i.API.Pin().IsPinned(ctx, p)

	// i.API.Routing().Provide(ctx, p, func(dps *options.DhtProvideSettings) error {
	// 	// dps.Recursive =
	// 	return nil
	// })

	// ps, err := i.API.Routing().FindProviders(ctx, p, func(dfps *options.DhtFindProvidersSettings) error {
	// 	// dfps.NumProviders =
	// 	return nil
	// })
	// if err != nil {
	// 	return err
	// }

	// for info := range ps {
	// 	EmitPeerFound{
	// 		Emitter: e,
	// 	}.HandlePeerFound(info)
	// }

	<-ctx.Done()
	return ctx.Err()
}
