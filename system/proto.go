package system

import (
	"fmt"

	"github.com/mr-tron/base58/base58"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	// TODO:  go back and pick good values for these
	P_WW = 9001 + iota
	P_PID
	P_METHOD
	P_STACK
	P_ARGS
)

var (
	protoWW = ma.Protocol{
		Name:       "ww",
		Code:       P_WW,
		VCode:      ma.CodeToVarint(P_WW),
		Size:       ma.LengthPrefixedVarSize, // /ww/0.1.0
		Transcoder: TextTranscoder{},
	}
	protoPID = ma.Protocol{
		Name:       "pid",
		Code:       P_PID,
		VCode:      ma.CodeToVarint(P_PID),
		Size:       20, // 160bit PID -- /pid/8uUq39P2CPPMk8zFNpE5tjTm2ge
		Transcoder: PidTranscoder{},
	}
	protoMethod = ma.Protocol{
		Name:       "method",
		Code:       P_METHOD,
		VCode:      ma.CodeToVarint(P_METHOD),
		Size:       ma.LengthPrefixedVarSize, // /method/foo
		Transcoder: TextTranscoder{},
	}
	protoStack = ma.Protocol{
		Name:       "stack",
		Code:       P_STACK,
		VCode:      ma.CodeToVarint(P_STACK),
		Size:       ma.LengthPrefixedVarSize, // /stack/jTm2ge
		Transcoder: Base58Transcoder{},
	}
)

type TextTranscoder struct{}

// Validates and encodes to bytes a multiaddr that's in the string representation.
func (t TextTranscoder) StringToBytes(s string) ([]byte, error) {
	return []byte(s), nil
}

// Validates and decodes to a string a multiaddr that's in the bytes representation.
func (t TextTranscoder) BytesToString(b []byte) (string, error) {
	return string(b), nil
}

// Validates bytes when parsing a multiaddr that's already in the bytes representation.
func (t TextTranscoder) ValidateBytes(b []byte) error {
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

type Base58Transcoder struct{}

// Validates and encodes to bytes a multiaddr that's in the string representation.
func (t Base58Transcoder) StringToBytes(s string) ([]byte, error) {
	return base58.FastBase58Decoding(s)
}

// Validates and decodes to a string a multiaddr that's in the bytes representation.
func (t Base58Transcoder) BytesToString(b []byte) (string, error) {
	return base58.FastBase58Encoding(b), nil
}

// Validates bytes when parsing a multiaddr that's already in the bytes representation.
func (t Base58Transcoder) ValidateBytes(b []byte) error {
	return nil
}

func init() {
	for _, p := range []ma.Protocol{
		protoWW,
		protoPID,
		protoMethod,
		protoStack,
	} {
		if err := ma.AddProtocol(p); err != nil {
			panic(err)
		}
	}
}
