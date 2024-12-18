package boot

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/lthibault/jitterbug/v2"
	"github.com/multiformats/go-multihash"
	"github.com/wetware/go/system"
)

// type Bootstrapper interface {
// 	// Bootstrap allows callers to hint to the routing system to get into dht
// 	// Boostrapped state and remain there.
// 	Bootstrap(context.Context) error
// }

type DHT struct {
	NS  string
	Env *system.Env
	TTL time.Duration
}

func (dht DHT) String() string {
	return "routing"
}

func (dht DHT) Name() string {
	if dht.NS == "" {
		return "ww"
	}

	return dht.NS
}

func (dht DHT) Log() *slog.Logger {
	return dht.Env.Log().With(
		"service", dht.String(),
		"ns", dht.Name(),
		"ttl", dht.TTL)
}

func (dht DHT) MinTTL() time.Duration {
	return (dht.TTL * 3) / 4 // TTL * .75
}

func (dht DHT) Serve(ctx context.Context) error {
	if dht.TTL <= 0 {
		dht.TTL = time.Hour * 24
	}

	dht.Log().DebugContext(ctx, "service started")

	timer := jitterbug.New(dht.TTL, &jitterbug.Uniform{
		Min: dht.MinTTL(),
	})
	defer timer.Stop()

	for {
		if err := dht.Announce(ctx); err != nil {
			return err
		}
		dht.Log().DebugContext(ctx, "announced peer")

		select {
		case <-timer.C:
			dht.Log().DebugContext(ctx, "woke up")

		case <-ctx.Done():
			return nil
		}
	}
}

func (dht DHT) Announce(ctx context.Context) error {
	b, err := multihash.Encode([]byte(dht.Name()), multihash.BLAKE3)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	key := cid.NewCidV1(multihash.BLAKE3, b)
	return dht.Env.DHT.Provide(ctx, key, true)
}
