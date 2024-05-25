package vat

import (
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tetratelabs/wazero/api"
)

func ProtoFromRoot(path string) protocol.ID {
	path = filepath.Join(Proto, path)
	return protocol.ID(path)
}

func ProtoFromModule(mod api.Module) protocol.ID {
	return ProtoFromRoot(mod.Name())
}
