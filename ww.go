package ww

import (
	"context"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/thejerf/suture/v4"
	"github.com/wetware/ww/boot"
	"github.com/wetware/ww/util"
)

const Proto = "/ww/0.0.0"

type NetworkBehavior interface {
	OnLocalAddrsUpdated(context.Context, event.EvtLocalAddressesUpdated)
	OnPeerFound(context.Context, boot.EvtPeerFound)
}

var events = []any{
	// libp2p events
	new(event.EvtLocalAddressesUpdated),
	// new(event.EvtLocalProtocolsUpdated),
	// new(event.EvtLocalReachabilityChanged),
	// new(event.EvtNATDeviceTypeChanged),
	// new(event.EvtPeerConnectednessChanged),
	// new(event.EvtPeerIdentificationCompleted),
	// new(event.EvtPeerIdentificationFailed),
	// new(event.EvtPeerProtocolsUpdated),

	// wetware events
	new(boot.EvtPeerFound),
}

type EventLoop struct {
	Name     string
	Bus      event.Bus
	Behavior NetworkBehavior
	Services []suture.Service
}

var _ suture.Service = (*EventLoop)(nil)

func (loop EventLoop) Serve(ctx context.Context) error {
	sub, err := loop.Bus.Subscribe(events)
	if err != nil {
		return err
	}
	defer sub.Close()

	root := loop.NewSupervisor()
	for _, s := range loop.Services {
		root.Add(s)
	}
	errs := root.ServeBackground(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-errs:
			return err

		case v := <-sub.Out():
			switch e := v.(type) {
			case event.EvtLocalAddressesUpdated:
				loop.Behavior.OnLocalAddrsUpdated(ctx, e)

			case boot.EvtPeerFound:
				loop.Behavior.OnPeerFound(ctx, e)

			default:
				panic(UnhandledEvent{e})
			}
		}
	}
}

func (loop EventLoop) NewSupervisor() *suture.Supervisor {
	return suture.New(loop.Name, suture.Spec{
		EventHook: util.EventHook,
	})
}
