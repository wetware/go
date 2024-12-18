package run

// import (
// 	"bytes"
// 	"io"
// 	"net/http"
// 	"os"

// 	"github.com/hashicorp/go-memdb"
// 	"github.com/ipfs/kubo/client/rpc"
// 	iface "github.com/ipfs/kubo/core/coreiface"
// 	"github.com/libp2p/go-libp2p"
// 	"github.com/libp2p/go-libp2p-kad-dht/dual"
// 	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
// 	ma "github.com/multiformats/go-multiaddr"
// 	"github.com/tetratelabs/wazero"
// 	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
// 	"github.com/urfave/cli/v2"
// 	ww "github.com/wetware/go"
// 	"github.com/wetware/go/boot"
// 	"github.com/wetware/go/system"
// )

// func Command() *cli.Command {
// 	return &cli.Command{
// 		Name: "run",
// 		Flags: []cli.Flag{
// 			&cli.StringFlag{
// 				Name:    "ipfs",
// 				EnvVars: []string{"WW_IPFS"},
// 				Usage:   "multi`addr` of IPFS node, or \"local\"",
// 				Value:   "local",
// 			},
// 			&cli.StringSliceFlag{
// 				Name:    "env",
// 				EnvVars: []string{"WW_ENV"},
// 			},
// 			&cli.BoolFlag{
// 				Name:    "wasm-debug",
// 				EnvVars: []string{"WW_WASM_DEBUG"},
// 				Usage:   "enable wasm debug symbols",
// 			},
// 		},
// 		Action: run(),
// 	}
// }

// func run() cli.ActionFunc {
// 	return func(c *cli.Context) error {
// 		ipfs, err := newIPFSClient(c)
// 		if err != nil {
// 			return err
// 		}

// 		// Set up libp2p host and DHT
// 		////
// 		h, err := libp2p.New()
// 		if err != nil {
// 			return err
// 		}
// 		defer h.Close()

// 		dht, err := dual.New(c.Context, h)
// 		if err != nil {
// 			return err
// 		}
// 		defer dht.Close()

// 		h = routedhost.Wrap(h, dht)

// 		// Start a multicast DNS service that searches for local
// 		// peers in the background.
// 		////
// 		d, err := boot.MDNS{
// 			Host: h,
// 			Handler: boot.PeerHandler{
// 				Peerstore:    h.Peerstore(),
// 				Bootstrapper: dht,
// 			},
// 		}.New()
// 		if err != nil {
// 			return err
// 		}
// 		defer d.Close()

// 		r := wazero.NewRuntimeWithConfig(c.Context, wazero.NewRuntimeConfig().
// 			WithDebugInfoEnabled(c.Bool("debug")).
// 			WithCloseOnContextDone(true))

// 		wasi, err := wasi_snapshot_preview1.Instantiate(c.Context, r)
// 		if err != nil {
// 			return err
// 		}
// 		defer wasi.Close(c.Context)

// 		db, err := memdb.NewMemDB(&system.Schema)
// 		if err != nil {
// 			return err
// 		}

// 		return ww.Env{
// 			IPFS: ipfs,
// 			Host: h,
// 			Cmd: system.Cmd{
// 				Stdin:  stdin(c),
// 				Args:   c.Args().Slice(),
// 				Env:    c.StringSlice("env"),
// 				Stdout: c.App.Writer,
// 				Stderr: c.App.ErrWriter},
// 			Net: system.Net{
// 				Host:   h,
// 				Router: db},
// 			FS: system.Anchor{
// 				Ctx:  c.Context,
// 				Host: h,
// 				IPFS: ipfs},
// 		}.Bind(c.Context, r)
// 	}
// }

// func newIPFSClient(c *cli.Context) (ipfs iface.CoreAPI, err error) {
// 	var a ma.Multiaddr
// 	if s := c.String("ipfs"); s == "local" {
// 		ipfs, err = rpc.NewLocalApi()
// 	} else if a, err = ma.NewMultiaddr(s); err == nil {
// 		ipfs, err = rpc.NewApiWithClient(a, http.DefaultClient)
// 	}

// 	return
// }

// func stdin(c *cli.Context) io.Reader {
// 	switch r := c.App.Reader.(type) {
// 	case *os.File:
// 		info, err := r.Stat()
// 		if err != nil {
// 			panic(err)
// 		}

// 		if info.Size() <= 0 {
// 			break
// 		}

// 		return &io.LimitedReader{
// 			R: c.App.Reader,
// 			N: 1<<32 - 1, // max u32
// 		}
// 	}

// 	return &bytes.Reader{} // empty buffer
// }
