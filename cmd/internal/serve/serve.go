package serve

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/event"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/tetratelabs/wazero"
	"github.com/thejerf/suture/v4"
	"github.com/urfave/cli/v2"
	ww "github.com/wetware/go"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
	"github.com/wetware/go/util"
)

func Command() *cli.Command {
	return &cli.Command{
		Name: "serve",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ipfs",
				EnvVars: []string{"WW_IPFS"},
				Usage:   "multi`addr` of IPFS node, or \"local\"",
				Value:   "local",
			},
			&cli.StringSliceFlag{
				Name:    "env",
				EnvVars: []string{"WW_ENV"},
			},
			&cli.BoolFlag{
				Name:    "wasm-debug",
				EnvVars: []string{"WW_WASM_DEBUG"},
				Usage:   "enable wasm debug symbols",
			},
		},
		Action: func(c *cli.Context) error {
			ipfs, err := newIPFSClient(c)
			if err != nil {
				return err
			}

			h, err := libp2p.New()
			if err != nil {
				return err
			}
			defer h.Close()

			s := suture.New("ww", suture.Spec{
				EventHook: util.EventHook,
			})

			// Start a multicast DNS service that searches for local
			// peers in the background.
			s.Add(ww.MDNS{
				Host:    h,
				Handler: ww.StorePeer{Peerstore: h.Peerstore()},
			})

			// Global compilation cache
			cache := wazero.NewCompilationCache()
			defer cache.Close(c.Context)

			s.Add(ww.Server{
				IPFS: ipfs,
				Host: h,
				Env: ww.Env{
					IO: system.IO{
						Args:   c.Args().Slice(),
						Env:    c.StringSlice("env"),
						Stdin:  stdin(c),
						Stdout: c.App.Writer,
						Stderr: c.App.ErrWriter,
					},
					Net: system.Net{
						Proto: ww.Proto,
						Host:  h,
						Handler: system.HandlerFunc(func(ctx context.Context, p *proc.P) error {
							slog.InfoContext(ctx, "process started",
								"peer", h.ID(),
								"proc", p.String())
							defer slog.WarnContext(ctx, "process stopped",
								"peer", h.ID(),
								"proc", p.String())
							<-ctx.Done()
							return ctx.Err()
						}),
					},
					FS: system.FS{
						Ctx:  c.Context,
						Host: h,
						IPFS: ipfs,
					},
				},
				RuntimeConfig: wazero.NewRuntimeConfig().
					WithCompilationCache(cache).
					WithDebugInfoEnabled(c.Bool("debug")).
					WithCloseOnContextDone(true),
			})

			sub, err := h.EventBus().Subscribe([]any{
				new(event.EvtLocalAddressesUpdated)})
			if err != nil {
				return err
			}
			defer sub.Close()

			done := s.ServeBackground(c.Context)
			for {
				var v any
				select {
				case err := <-done:
					return err // exit
				case v = <-sub.Out():
					// event received
				}

				// Synchronous event handler
				switch ev := v.(type) {
				case *event.EvtLocalAddressesUpdated:
					// TODO(easy):  emit to libp2p topic
					slog.InfoContext(c.Context, "local addresses updated",
						"peer", h.ID(),
						"current", ev.Current,
						"removed", ev.Removed,
						"diffs", ev.Diffs)

				case *event.EvtLocalProtocolsUpdated:
					// TODO(easy):  emit to libp2p topic
					slog.InfoContext(c.Context, "local protocols updated",
						"peer", h.ID(),
						"added", ev.Added,
						"removed", ev.Removed)
				}

			}
		},
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
	case interface{ Len() int }:
		if r.Len() <= 0 {
			break
		}

		return &io.LimitedReader{
			R: c.App.Reader,
			N: min(int64(r.Len()), 1<<32-1), // max u32
		}

	case interface{ Stat() (fs.FileInfo, error) }:
		info, err := r.Stat()
		if err != nil {
			slog.Error("failed to get file info for stdin",
				"reason", err)
			break
		} else if info.Size() <= 0 {
			break
		}

		return &io.LimitedReader{
			R: c.App.Reader,
			N: min(info.Size(), 1<<32-1), // max u32
		}
	}

	return &bytes.Reader{} // empty buffer
}
