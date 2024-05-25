package vat

import (
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tetratelabs/wazero/api"
)

func ProtoFromModule(mod api.Module) protocol.ID {
	path := filepath.Join(Proto, mod.Name())
	return protocol.ID(path)
}
