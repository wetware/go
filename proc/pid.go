package proc

import (
	"crypto/rand"
	"io"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mr-tron/base58/base58"
)

type PID [20]byte // 160bit opaque identifier

func NewPID() (pid PID) {
	var err error
	if pid, err = ReadPID(rand.Reader); err != nil {
		panic(err)
	}

	return
}

func ReadPID(r io.Reader) (pid PID, err error) {
	var n int // if no error and don't read 20 bytes, sound the alarm.
	if n, err = r.Read(pid[:]); n != len(pid) && err == nil {
		err = io.ErrUnexpectedEOF
	}

	return
}

func ParsePID(s string) (pid PID, err error) {
	var buf []byte
	if buf, err = base58.FastBase58Decoding(s); err == nil {
		copy(pid[:], buf)
	}
	return
}

func (pid PID) String() string {
	return base58.FastBase58Encoding(pid[:])
}

func (pid PID) Proto() protocol.ID {
	proto := pid.String()
	return protocol.ID(proto)
}
