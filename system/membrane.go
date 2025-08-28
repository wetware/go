package system

import (
	"context"
	"fmt"

	capnp "capnproto.org/go/capnp/v3"
)

var _ Importer_Server = (*Membrane)(nil)

type ServiceToken [20]byte // 20 bytes = 160 bits = 20 hex characters

type Membrane map[ServiceToken]capnp.Client

func (m Membrane) Import(ctx context.Context, call Importer_import) error {
	raw, err := call.Args().Envelope()
	if err != nil {
		return err
	}

	var token ServiceToken
	if copy(token[:], raw) != len(raw) {
		return fmt.Errorf("invalid service token length: got %d, want %d", len(raw), len(token))
	}

	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	return res.SetService(m[token])
}
