package system

import (
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mr-tron/base58/base58"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	// TODO:  go back and pick good values for these
	P_WW = 9001 + iota
	P_PID
)

var (
	protoWW = ma.Protocol{
		Name:       "ww",
		Code:       P_WW,
		VCode:      ma.CodeToVarint(P_WW),
		Size:       ma.LengthPrefixedVarSize, // ww/<semver>
		Transcoder: WwTranscoder{},
	}
	protoPID = ma.Protocol{
		Name:       "pid",
		Code:       P_PID,
		VCode:      ma.CodeToVarint(P_PID),
		Size:       20, // 160bit PID
		Transcoder: PidTranscoder{},
	}
)

type WwTranscoder struct{}

// Validates and encodes to bytes a multiaddr that's in the string representation.
func (t WwTranscoder) StringToBytes(s string) ([]byte, error) {
	return []byte(s), nil
}

// Validates and decodes to a string a multiaddr that's in the bytes representation.
func (t WwTranscoder) BytesToString(b []byte) (string, error) {
	return string(b), nil
}

// Validates bytes when parsing a multiaddr that's already in the bytes representation.
func (t WwTranscoder) ValidateBytes(b []byte) error {
	return nil
}

type PidTranscoder struct{}

// Validates and encodes to bytes a multiaddr that's in the string representation.
func (t PidTranscoder) StringToBytes(s string) ([]byte, error) {
	return base58.FastBase58Decoding(s)
}

// Validates and decodes to a string a multiaddr that's in the bytes representation.
func (t PidTranscoder) BytesToString(b []byte) (string, error) {
	return base58.FastBase58Encoding(b), nil
}

// Validates bytes when parsing a multiaddr that's already in the bytes representation.
func (t PidTranscoder) ValidateBytes(b []byte) error {
	if size := len(b); size != 20 {
		return fmt.Errorf("expected 20byte PID, got %d", size)
	}

	return nil
}

func init() {
	for _, p := range []ma.Protocol{
		protoWW,
		protoPID,
	} {
		if err := ma.AddProtocol(p); err != nil {
			panic(err)
		}
	}
}

type Path struct {
	ma.Multiaddr
}

func NewPath(name string) (p Path, err error) {
	p.Multiaddr, err = ma.NewMultiaddr(name)
	return
}

func (p Path) Version() (semver.Version, error) {
	s, err := p.ValueForProtocol(P_WW)
	if err != nil {
		return semver.Version{}, err
	}

	return semver.Parse(s)
}

func (p Path) Peer() (peer.ID, error) {
	id, err := p.ValueForProtocol(ma.P_P2P)
	if err != nil {
		return "", err
	}

	return peer.Decode(id)
}

func (p Path) Proto() (protocol.ID, error) {
	s, err := p.ValueForProtocol(P_PID)
	if err != nil {
		return "", err
	}

	proto := "/proc/" + s
	return protocol.ID(proto), nil
}
