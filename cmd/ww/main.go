package main

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/lmittmann/tint"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
	syncutils "github.com/wetware/go/util/sync"
	"go.uber.org/multierr"

	"github.com/wetware/go/cmd/internal/serve"
	"github.com/wetware/go/system"
)

var env system.Env

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill)
	defer cancel()

	app := &cli.App{
		Name:           "wetware",
		Copyright:      "2020 The Wetware Project",
		Before:         setup,
		After:          teardown,
		DefaultCommand: "serve",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ns",
				EnvVars: []string{"WW_NS"},
				Value:   "ww",
				Usage:   "cluster namespace",
			},
			&cli.StringSliceFlag{
				Name:    "dial",
				EnvVars: []string{"WW_DIAL"},
				Aliases: []string{"d"},
				Usage:   "peer addr to dial",
			},
			&cli.StringSliceFlag{
				Name:    "listen",
				EnvVars: []string{"WW_LISTEN"},
				Aliases: []string{"l"},
				Usage:   "multiaddr to listen on",
			},
			&cli.BoolFlag{
				Name:    "json",
				EnvVars: []string{"WW_JSON"},
				Usage:   "output json logs",
			},
			&cli.StringFlag{
				Name:    "loglvl",
				EnvVars: []string{"WW_LOGLVL"},
				Value:   "info",
				Usage:   "logging level: debug, info, warn, error",
			},
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Usage:   "multi`addr` of IPFS node, or \"local\"",
				Value:   "local",
			},
		},
		Commands: []*cli.Command{
			serve.Command(&env),
			// export.Command(&env),
			// run.Command(&env),
		},
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}

func setup(c *cli.Context) (err error) {
	log := slog.New(logger(c)).With(
		"version", system.Proto.Version)
	slog.SetDefault(log)

	// Set up IPFS and libp2p
	////
	if env.IPFS, err = newIPFSClient(c); err != nil {
		return
	} else if env.Host, err = newLibp2pHost(c); err != nil {
		return
	}

	// Initialize a Dual DHT, which maintains two separate Kademlia routing tables:
	// one for the Wide Area Network (WAN) and another for the Local Area Network
	// (LAN). This separation allows for efficient routing in both local and global
	// contexts.
	//
	// The WAN DHT is bootstrapped with public bootstrap peers - well-known nodes
	// that help new peers join the network by providing initial routing table
	// entries. These peers are typically operated by IPFS infrastructure providers
	// and are accessible from anywhere on the internet.
	//
	// The LAN DHT is bootstrapped with private bootstrap peers - nodes that are
	// only accessible within the local network. This enables peer discovery and
	// content routing to work efficiently in isolated/local network environments,
	// without having to rely on public infrastructure.
	//
	// Both DHTs shuffle their bootstrap peers to prevent hotspots and ensure even
	// load distribution across the bootstrap nodes. User-provided bootstrap addresses
	// (-addr flag) are added to both DHTs to support custom network topologies.
	////
	env.DHT, err = dual.New(c.Context, env.Host,
		dual.WanDHTOption(dht.BootstrapPeersFunc(func() []peer.AddrInfo {
			public := append(
				env.PublicBootstrapPeers(),
				dht.GetDefaultBootstrapPeerAddrInfos()...)
			rand.Shuffle(len(public), func(i, j int) {
				public[i], public[j] = public[j], public[i]
			})

			return append(addrs(c), public...)
		})),
		dual.LanDHTOption(dht.BootstrapPeersFunc(func() []peer.AddrInfo {
			private := env.PrivateBootstrapPeers()
			rand.Shuffle(len(private), func(i, j int) {
				private[i], private[j] = private[j], private[i]
			})

			return append(addrs(c), private...)
		})))
	if err != nil {
		return
	}

	// Wrap the host in a routed host, which intercepts all network operations
	// and uses the DHT for peer routing. This enables automatic peer discovery
	// and routing through the DHT when direct connections aren't available.
	// When the host attempts to dial a peer, the routed host will first check
	// if it has a direct connection. If not, it will query the DHT to find
	// the peer's addresses before attempting the connection.
	////
	env.Host = routedhost.Wrap(env.Host, env.DHT)
	env.Host.Peerstore().AddAddrs(
		env.Host.ID(),
		env.Host.Addrs(),
		peerstore.PermanentAddrTTL)

	// Set cluster namespace
	////
	env.NS = c.String("ns")

	return bootstrap(c)
}

func bootstrap(c *cli.Context) error {
	// Concurrently attempt to connect to all provided bootstrap addresses.
	// Each connection attempt is run in its own goroutine. If any connection
	// fails, the error is stored in the atomic error value. The wait group
	// ensures we wait for all connection attempts to complete before proceeding.
	////
	var join syncutils.Any // join strategy:  any successful connection => ok
	for _, info := range addrs(c) {
		join.Go(func() error {
			return env.Host.Connect(c.Context, info)
		})
	}

	if err := join.Wait(); err != nil {
		return err
	}

	return env.DHT.Bootstrap(c.Context)
}

func newLibp2pHost(c *cli.Context) (host.Host, error) {
	listenAddrs, err := listenAddrs(c)
	if err != nil {
		return nil, err
	}

	return libp2p.New(
		libp2p.ListenAddrs(listenAddrs...),
	)
}

func listenAddrs(c *cli.Context) ([]ma.Multiaddr, error) {
	var addrs []ma.Multiaddr
	var errs []error
	for _, a := range c.StringSlice("listen") {
		if m, err := ma.NewMultiaddr(a); err != nil {
			errs = append(errs, err)
		} else {
			addrs = append(addrs, m)
		}
	}

	return addrs, multierr.Combine(errs...)
}

// addrs returns bootstrap addresses parsed from args
func addrs(c *cli.Context) []peer.AddrInfo {
	ps := map[peer.ID][]ma.Multiaddr{}
	for _, a := range c.StringSlice("dial") {
		m, err := ma.NewMultiaddr(a)
		if err != nil {
			slog.Debug("failed to parse multiaddr",
				"addr", a,
				"reason", err)
			continue
		}

		s, err := m.ValueForProtocol(ma.P_P2P)
		if err != nil {
			slog.Debug("failed to parse value for protocol",
				"proto", "p2p",
				"addr", a,
				"reason", err)
			continue
		}
		id, err := peer.Decode(s)
		if err != nil {
			slog.Debug("failed to decode peer ID",
				"addr", a,
				"reason", err)
			continue
		}

		addr := m.Decapsulate(ma.StringCast("/p2p/" + s))
		ps[id] = append(ps[id], addr)
	}

	var dial []peer.AddrInfo
	for id, addrs := range ps {
		if len(addrs) > 0 {
			dial = append(dial, peer.AddrInfo{
				ID:    id,
				Addrs: addrs,
			})
		}
	}

	return dial
}

func teardown(c *cli.Context) error {
	// host was started?
	if env.Host != nil {
		return env.Host.Close()
	}

	return nil
}

func logger(c *cli.Context) slog.Handler {
	// For robots?
	if c.Bool("json") {
		return slog.NewJSONHandler(c.App.ErrWriter, &slog.HandlerOptions{
			Level: loglvl(c),
		})
	}

	// For people
	return tint.NewHandler(c.App.ErrWriter, &tint.Options{
		Level:      loglvl(c),
		TimeFormat: time.Kitchen,
	})
}

func loglvl(c *cli.Context) slog.Leveler {
	switch c.String("loglvl") {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}

	return slog.LevelInfo
}

func newIPFSClient(c *cli.Context) (iface.CoreAPI, error) {
	if !c.IsSet("ipfs") {
		return rpc.NewLocalApi()
	}

	a, err := ma.NewMultiaddr(c.String("ipfs"))
	if err != nil {
		return nil, err
	}

	return rpc.NewApiWithClient(a, http.DefaultClient)
}
