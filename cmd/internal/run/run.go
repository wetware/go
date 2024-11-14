package run

import (
	"io"
	"net/http"

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

		r := wazero.NewRuntimeWithConfig(c.Context, wazero.NewRuntimeConfig().
			// WithCompilationCache().
			WithDebugInfoEnabled(c.Bool("debug")).
			WithCloseOnContextDone(true))
		defer r.Close(c.Context)

		root, err := path.NewPath(c.Args().First())
		if err != nil {
			return err
		}

		unixFS := system.IPFS{
			Ctx:  c.Context,
			Unix: ipfs.Unixfs(),
		}

		return ww.Env{
			Args:    c.Args().Slice(),
			Vars:    c.StringSlice("env"),
			Stdin:   maybeStdin(c),
			Stdout:  maybeStdout(c),
			Stderr:  maybeStderr(c),
			Host:    h,
			Runtime: r,
			Root:    root.String(),
			FS:      unixFS,
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

func maybeStdin(c *cli.Context) io.Reader {
	if c.Bool("interactive") {
		return c.App.Reader // TODO:  wrap with libreadline
	}

	return nil
}

func maybeStdout(c *cli.Context) io.Writer {
	if c.Bool("interactive") {
		return c.App.Writer
	}

	return io.Discard // TODO:  handle stdout
}

func maybeStderr(c *cli.Context) io.Writer {
	if c.Bool("interactive") {
		return c.App.ErrWriter
	}

	return io.Discard // TODO:  handle stderr
}
