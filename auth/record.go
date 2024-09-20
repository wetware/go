package auth

import (
	"bytes"
	"fmt"

	"github.com/libp2p/go-libp2p/core/record"
)

// AuthDomain_Nonce is the domain string used for peer records contained in an Envelope.
const AuthDomain_Nonce = "ww/auth/nonce"
const nonceSize = 20 // 160 bits

// PeerRecordEnvelopePayloadType is the type hint used to identify peer records in an Envelope.
// Defined in https://github.com/multiformats/multicodec/blob/master/table.csv
// with name "libp2p-peer-record".
var AuthDomain_PayloadType = []byte{0xbb, 0xbb} // TODO:  pick better numbers

var _ record.Record = (*expect)(nil)

type expect []byte

// Domain is the "signature domain" used when signing and verifying a particular
// Record type. The Domain string should be unique to your Record type, and all
// instances of the Record type must have the same Domain string.
func (e expect) Domain() string {
	return AuthDomain_Nonce
}

// Codec is a binary identifier for this type of record, ideally a registered multicodec
// (see https://github.com/multiformats/multicodec).
// When a Record is put into an Envelope (see record.Seal), the Codec value will be used
// as the Envelope's PayloadType. When the Envelope is later unsealed, the PayloadType
// will be used to look up the correct Record type to unmarshal the Envelope payload into.
func (e expect) Codec() []byte {
	return AuthDomain_PayloadType
}

// MarshalRecord converts a Record instance to a []byte, so that it can be used as an
// Envelope payload.
func (e expect) MarshalRecord() ([]byte, error) {
	if len(e) != nonceSize {
		return nil, fmt.Errorf("expected 160 bit nonce, got %d", bitlen(e))
	}

	return e, nil
}

// UnmarshalRecord unmarshals a []byte payload into an instance of a particular Record type.
func (e expect) UnmarshalRecord(b []byte) error {
	if len(b) != nonceSize {
		return fmt.Errorf("expected 160 bit nonce, got %d", bitlen(b))
	}

	if copy(e, b) != nonceSize {
		panic(e) // unreachable
	}

	if bytes.Equal(e, b) {
		return nil
	}

	return fmt.Errorf("expected nonce %x, got %x", e, b)
}

func bitlen(b []byte) int {
	return len(b) * 8
}
