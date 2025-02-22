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
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/lmittmann/tint"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"

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
				Name:    "addr",
				EnvVars: []string{"WW_ADDRS"},
				Aliases: []string{"a"},
				Usage:   "peer addr to dial",
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
	} else if env.Host, err = libp2p.New(); err != nil {
		return
	}

	env.DHT, err = dual.New(c.Context, env.Host,
		dual.WanDHTOption(dht.BootstrapPeersFunc(func() []peer.AddrInfo {
			public := env.PublicBootstrapPeers()
			rand.Shuffle(len(public), func(i, j int) {
				public[i], public[j] = public[j], public[i]
			})

			args := addrs(c)
			return append(args, public...)
		})),
		dual.LanDHTOption(dht.BootstrapPeersFunc(func() []peer.AddrInfo {
			private := env.PrivateBootstrapPeers()
			rand.Shuffle(len(private), func(i, j int) {
				private[i], private[j] = private[j], private[i]
			})

			args := addrs(c)
			return append(args, private...)
		})))
	if err != nil {
		return
	}
	env.Host = routedhost.Wrap(env.Host, env.DHT)
	env.Host.Peerstore().AddAddrs(
		env.Host.ID(),
		env.Host.Addrs(),
		peerstore.PermanentAddrTTL)

	// Set cluster namespace
	////
	env.NS = c.String("ns")

	return
}

// addrs returns bootstrap addresses parsed from args
func addrs(c *cli.Context) []peer.AddrInfo {
	ps := map[peer.ID][]ma.Multiaddr{}
	for _, a := range c.StringSlice("addr") {
		m, err := ma.NewMultiaddr(a)
		if err != nil {
			// ...
		}

		s, err := m.ValueForProtocol(ma.P_P2P)
		if err != nil {
			// ...
		}
		id, err := peer.Decode(s)
		if err != nil {
			// ...
		}

		addr := m.Decapsulate(ma.StringCast("p2p/" + s))
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
