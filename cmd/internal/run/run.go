package run

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/urfave/cli/v2"
	ww "github.com/wetware/go"
	"github.com/wetware/go/proc"
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
				Name:  "stdin",
				Usage: "bind stdin to wasm guest",
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
		peerHook := system.StorePeer{Peerstore: h.Peerstore()}
		d := mdns.NewMdnsService(h, "ww.local", peerHook)
		if err := d.Start(); err != nil {
			return err
		}
		defer d.Close()

		// Set up WASM runtime and host modules
		r := wazero.NewRuntimeWithConfig(c.Context, wazero.NewRuntimeConfig().
			// WithCompilationCache().
			WithDebugInfoEnabled(c.Bool("debug")).
			WithCloseOnContextDone(true))
		defer r.Close(c.Context)

		wasi, err := wasi_snapshot_preview1.Instantiate(c.Context, r)
		if err != nil {
			return err
		}
		defer wasi.Close(c.Context)

		return ww.Env{
			IO: system.IO{
				Args:   c.Args().Slice(),
				Env:    c.StringSlice("env"),
				Stdin:  stdin(c),
				Stdout: stdout(c),
				Stderr: stderr(c),
			},
			Net: system.Net{
				Host:    h,
				Handler: handler(c, h),
			},
			FS: system.IPFS{
				Ctx:  c.Context,
				Unix: ipfs.Unixfs(),
			},
		}.Bind(c.Context, r)
	}
}

func handler(c *cli.Context, h host.Host) system.HandlerFunc {
	return func(ctx context.Context, p *proc.P) error {
		sub, err := h.EventBus().Subscribe([]any{
			new(event.EvtLocalAddressesUpdated),
			new(event.EvtLocalProtocolsUpdated)})
		if err != nil {
			return err
		}
		defer sub.Close()

		// asynchronous event loop
		slog.InfoContext(ctx, "wetware started",
			"peer", h.ID(),
			"path", c.Args().First(),
			"proc", p.String(),
			"proto", ww.Proto.String())
		defer slog.WarnContext(ctx, "wetware stopped",
			"peer", h.ID(),
			"path", c.Args().First(),
			"proc", p.String(),
			"proto", ww.Proto.String())

		for {
			var v any
			select {
			case <-ctx.Done():
				return ctx.Err()

			case v = <-sub.Out():
				// current event is assigned to v

				switch ev := v.(type) {
				case *event.EvtLocalAddressesUpdated:
					// TODO(easy):  emit to libp2p topic
					slog.InfoContext(ctx, "local addresses updated",
						"peer", h.ID(),
						"current", ev.Current,
						"removed", ev.Removed,
						"diffs", ev.Diffs)

				case *event.EvtLocalProtocolsUpdated:
					// TODO(easy):  emit to libp2p topic
					slog.InfoContext(ctx, "local protocols updated",
						"peer", h.ID(),
						"added", ev.Added,
						"removed", ev.Removed)
				}
			}
		}
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

func stdin(c *cli.Context) io.Reader {
	switch r := c.App.Reader.(type) {
	case *os.File:
		info, err := r.Stat()
		if err != nil {
			panic(err)
		}

		if info.Size() <= 0 {
			break
		}

		return &io.LimitedReader{
			R: c.App.Reader,
			N: 1<<32 - 1, // max u32
		}
	}

	return &bytes.Reader{} // empty buffer
}

func stdout(c *cli.Context) io.Writer {
	return c.App.Writer
}

func stderr(c *cli.Context) io.Writer {
	return c.App.ErrWriter
}
