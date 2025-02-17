package glia

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"golang.org/x/sync/errgroup"

	"github.com/wetware/go/boot"
	"github.com/wetware/go/system"
)

var ErrU16Overflow = errors.New("varint overflows u16")

type P2P struct {
	Env    *system.Env
	Router Router
	Boot   discovery.Discoverer
}

func (p2p P2P) Log() *slog.Logger {
	return p2p.Env.Log().With(
		"service", p2p.String())
}

func (p2p P2P) String() string {
	return "p2p"
}

func (p2p P2P) Serve(ctx context.Context) error {
	env := p2p.Env
	proto := system.Proto.Unwrap()
	env.Host.SetStreamHandlerMatch(proto,
		func(id protocol.ID) bool {
			root := system.Proto.Path()
			return strings.HasPrefix(string(id), root)
		},
		func(s network.Stream) {
			defer s.Close()

			if dl, ok := ctx.Deadline(); ok {
				if err := s.SetDeadline(dl); err != nil {
					p2p.Log().WarnContext(ctx, "failed to set deadline",
						"reason", err)
					// non-fatal; continue along...
				}
			}

			if err := p2p.ServeStream(ctx, P2PStream{Stream: s}); err != nil {
				p2p.Log().ErrorContext(ctx, "failed to serve stream",
					"reason", err,
					"stream", s.ID())
			}
		})
	defer env.Host.RemoveStreamHandler(proto)
	p2p.Log().DebugContext(ctx, "service started")

	if err := p2p.Bootstrap(ctx); err != nil {
		return err
	}

	// If we made it this far, we're bootstrapped.  Wait for shutdown
	// signal.
	<-ctx.Done()
	return ctx.Err()
}

func (p2p P2P) Bootstrap(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	if p2p.Boot == nil {
		addrs, err := p2p.IPFSPeers(ctx)
		if err != nil {
			return fmt.Errorf("get ipfs peers: %w", err)
		}
		p2p.Boot = addrs
	}

	peers, err := p2p.Boot.FindPeers(ctx, p2p.Env.NS, discovery.Limit(8))
	if err != nil {
		return err
	}

	var ok atomic.Bool
	var g errgroup.Group
	for info := range peers {
		g.Go(func(info peer.AddrInfo) func() error {
			return func() error {
				err := p2p.Env.Host.Connect(ctx, info)
				if err == nil {
					ok.Store(true)
				}
				return err
			}
		}(info))
	}

	// Wait for all threads to finish.  Fail if none succeeded.
	////
	if err := g.Wait(); !ok.Load() {
		return err
	}

	// At least one connection succeeded.  Onward!
	return p2p.Env.DHT.Bootstrap(ctx)
}

func (p2p P2P) IPFSPeers(ctx context.Context) (boot.StaticAddrs, error) {
	// Get peers from IPFS swarm
	peers, err := p2p.Env.IPFS.Swarm().Peers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPFS peers: %w", err)
	}

	// Build map of peer IDs to their multiaddrs
	peerMap := make(map[peer.ID][]ma.Multiaddr)
	for _, p := range peers {
		id := p.ID()
		peerMap[id] = append(peerMap[id], p.Address())
	}

	// Convert map entries to AddrInfos
	addrs := make(boot.StaticAddrs, 0, len(peerMap))
	for id, maddrs := range peerMap {
		info := &peer.AddrInfo{
			ID:    id,
			Addrs: maddrs,
		}
		addrs = append(addrs, *info)
	}

	// Defensive measure:  randomize the order in which peers are
	// dialed in order to distribute the load.
	rand.Shuffle(len(addrs), func(i, j int) {
		addrs[i], addrs[j] = addrs[j], addrs[i]
	})

	return addrs, nil
}

func (p2p P2P) ServeStream(ctx context.Context, s Stream) error {
	defer s.Close()
	// Glia RPC is a synchronous RPC protocol models one round-trip
	// (request-response) between a server and a client.  The round-
	// trip models a synchronous method call on an object.
	////

	// Local call?
	////
	if p2p.Env.Host.ID().String() == s.Destination() {
		p, err := p2p.Router.GetProc(s.ProcID())
		if err != nil {
			return err
		}

		if err := p.Reserve(ctx, s); err != nil {
			return err
		}
		defer p.Release()

		return p.Method(s.MethodName()).CallWithStack(ctx, nil) // TODO:  stack
	}

	// Forward the call
	////
	dst, err := peer.Decode(s.Destination())
	if err != nil {
		return err
	}
	remote, err := p2p.Env.Host.NewStream(ctx, dst, s.Protocol())
	if err != nil {
		return err
	}
	defer s.Close()

	// Forward the request stream
	if _, err := io.Copy(remote, s); err != nil {
		return err
	}
	if err := remote.CloseWrite(); err != nil {
		return err
	}

	// Read back the response stream
	if _, err := io.Copy(s, remote); err != nil {
		return err
	}
	if err := remote.CloseRead(); err != nil {
		return err
	}

	return nil
}

type P2PStream struct {
	network.Stream
}

var _ Stream = (*P2PStream)(nil)

func (s P2PStream) Close() error {
	return s.Stream.Close()
}

func (s P2PStream) Destination() string {
	proto := s.Protocol()
	p := path.Dir(string(proto))
	p = path.Dir(p)
	return path.Base(p)
}

func (s P2PStream) ProcID() string {
	proto := s.Protocol()
	dir := path.Dir(string(proto))
	return path.Base(dir)
}

func (s P2PStream) MethodName() string {
	proto := s.Protocol()
	return path.Base(string(proto))

}
