package system

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/record"
)

var (
	_ Importer_Server = (*Membrane)(nil)
	_ record.Record   = (*ServiceToken)(nil)
)

type Membrane struct {
}

func (m *Membrane) Import(ctx context.Context, call Importer_import) error {
	raw, err := call.Args().Envelope()
	if err != nil {
		return err
	}

	var token ServiceToken
	e, err := record.ConsumeTypedEnvelope(raw, &token)
	if err != nil {
		return err
	}

	issuer, err := peer.IDFromPublicKey(e.PublicKey)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "TODO:  finish importer implementation", "issuer", issuer)

	return nil
}

type ServiceToken [24]byte

func (t ServiceToken) Domain() string {
	return "ww.system.membrane"
}

func (t ServiceToken) Codec() []byte {
	// Return the codec identifier for service records
	// See: https://github.com/libp2p/go-libp2p/blob/master/core/record/record.go
	// Codec identifiers should be a varint
	// 0x7777 = 30583 in decimal
	// As a varint this should be: 0xF7, 0xEF, 0x01
	return []byte{0xF7, 0xEF, 0x01} // Varint encoding of "ww" (0x7777)
}

func (t ServiceToken) MarshalRecord() ([]byte, error) {
	return t[:], nil
}

func (t ServiceToken) UnmarshalRecord(b []byte) error {
	if copy(t[:], b) != len(b) {
		return fmt.Errorf("invalid record length: got %d, want %d", len(b), len(t))
	}

	return nil
}
