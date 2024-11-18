package ww

import (
	"context"
	"path"

	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type Server struct {
	IPFS          iface.CoreAPI
	Host          host.Host
	Env           Env
	RuntimeConfig wazero.RuntimeConfig
}

func (s Server) String() string {
	peer := string(s.Host.ID())
	return path.Join("/p2p", peer, Proto.String())
}

func (s Server) Serve(ctx context.Context) error {
	if s.RuntimeConfig == nil {
		s.RuntimeConfig = wazero.NewRuntimeConfig().
			WithCloseOnContextDone(true)
	}

	// Set up WASM runtime and host modules
	r := wazero.NewRuntimeWithConfig(ctx, s.RuntimeConfig)
	defer r.Close(ctx)

	wasi, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		return err
	}
	defer wasi.Close(ctx)

	return s.Env.Bind(ctx, r)
}
