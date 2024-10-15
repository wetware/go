package run

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ipfs/boxo/path"
	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/tetratelabs/wazero"
	"github.com/urfave/cli/v2"
	ww "github.com/wetware/go"
	"github.com/wetware/go/system"
)

func Command() *cli.Command {
	return &cli.Command{
		Name: "run",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Usage:   "multi`addr` of IPFS node, or \"local\"",
				Value:   "local",
			},
			&cli.StringFlag{
				Name:    "dial",
				EnvVars: []string{"WW_DIAL"},
				// Usage:   "",
				// Value: "", // TODO:  default to /ipfs/<CID> pointing to shell
			},
			&cli.StringSliceFlag{
				Name:    "env",
				EnvVars: []string{"WW_ENV"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				EnvVars: []string{"WW_DEBUG"},
				Usage:   "enable WASM debug values",
			},
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				EnvVars: []string{"WW_INTERACTIVE"},
				Usage:   "bind to process stdio",
			},
		},
		Action: run(),
	}
}

func run() cli.ActionFunc {
	return func(c *cli.Context) error {
		ipfs, err := newIPFSClient(c)
		if err != nil {
			return err
		}

		h, err := libp2p.New()
		if err != nil {
			return err
		}
		defer h.Close()

		// Create an mDNS service to discover peers on the local network
		d := mdns.NewMdnsService(h, "ww.local", peerStorer{Peerstore: h.Peerstore()})
		if err := d.Start(); err != nil {
			return err
		}
		defer d.Close()

		return ww.Env{
			IPFS: ipfs,
			Host: h,
			Boot: unixFSNode{
				Name: c.Args().Get(1),
				Unix: ipfs.Unixfs()},
			WASM: wazero.NewRuntimeConfig().
				WithDebugInfoEnabled(c.Bool("debug")).
				WithCloseOnContextDone(true),
			Module: moduleConfig(c, system.FSConfig{
				IPFS: ipfs,
				Host: h}),
		}.Serve(c.Context)
	}
}

func newIPFSClient(c *cli.Context) (ipfs iface.CoreAPI, err error) {
	var a ma.Multiaddr
	if s := c.String("ipfs"); s == "local" {
		ipfs, err = rpc.NewLocalApi()
	} else if a, err = ma.NewMultiaddr(s); err == nil {
		ipfs, err = rpc.NewApiWithClient(a, http.DefaultClient)
	}

	return
}

type peerStorer struct {
	peerstore.Peerstore
}

func (s peerStorer) HandlePeerFound(info peer.AddrInfo) {
	for _, addr := range info.Addrs {
		s.AddAddr(info.ID, addr, peerstore.AddressTTL) // assume a dynamic environment
	}
}

type unixFSNode struct {
	Name string
	Unix iface.UnixfsAPI
}

func (u unixFSNode) Load(ctx context.Context) ([]byte, error) {
	p, err := path.NewPath(u.Name)
	if err != nil {
		return nil, err
	}

	n, err := u.Unix.Get(ctx, p)
	if err != nil {
		return nil, err
	}
	defer n.Close()

	return io.ReadAll(n.(io.Reader))
}

func moduleConfig(c *cli.Context, fs fs.FS) wazero.ModuleConfig {
	config := wazero.NewModuleConfig().
		WithArgs(c.Args().Tail()...).
		WithName(c.Args().Tail()[0]).
		WithStdin(maybeStdin(c)).
		WithStdout(maybeStdout(c)).
		WithStderr(maybeStderr(c)).
		WithFSConfig(wazero.NewFSConfig().
			WithFSMount(fs, "/p2p/").
			WithFSMount(fs, "/ipfs/").
			WithFSMount(fs, "/ipns/").
			WithFSMount(fs, "/ipld/"))
	return withEnvironment(c, config)
}

func withEnvironment(c *cli.Context, config wazero.ModuleConfig) wazero.ModuleConfig {
	version := fmt.Sprintf("WW_VERSION=%s", ww.Version)
	root := fmt.Sprintf("WW_ROOT=%s", c.Args().First())
	env := append(c.StringSlice("env"), version, root)

	for _, v := range env {
		if maybePair := strings.SplitN(v, "=", 2); len(maybePair) == 2 {
			config = config.WithEnv(maybePair[0], maybePair[1])
		} else {
			slog.DebugContext(c.Context, "ignoring invalid environment variable",
				"value", v)
		}
	}

	return config
}

func maybeStdin(c *cli.Context) io.Reader {
	if c.Bool("interactive") {
		return c.App.Reader
	}

	return nil
}

func maybeStdout(c *cli.Context) io.Writer {
	if c.Bool("interactive") {
		return c.App.Writer
	}

	return nil
}

func maybeStderr(c *cli.Context) io.Writer {
	if c.Bool("interactive") {
		return c.App.ErrWriter
	}

	return nil
}
