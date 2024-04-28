package boot

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

type ENS struct {
	Bus event.Bus
	TTL time.Duration
}

func (e ENS) String() string {
	return fmt.Sprintf("ENS{%s}", e.TTL)
}

func (e *ENS) Serve(ctx context.Context) error {
	if e.TTL <= 0 {
		e.TTL = peerstore.AddressTTL
	}

	<-ctx.Done()
	return ctx.Err()
}
