package proc

import (
	"crypto/rand"
	"encoding/hex"
	"io"
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

func (pid PID) String() string {
	return hex.EncodeToString(pid[:])
}
