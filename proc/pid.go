package proc

import (
	"crypto/rand"
	"fmt"
	"io"
	"path"
	"reflect"

	"github.com/hashicorp/go-memdb"
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

var _ memdb.Indexer = (*PIDIndexer)(nil)
var _ memdb.SingleIndexer = (*PIDIndexer)(nil)

type PIDIndexer struct{}

// FromArgs is called to build the exact index key from a list of arguments.
func (i PIDIndexer) FromArgs(args ...interface{}) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("must provide only a single argument")
	}

	switch arg0 := args[0].(type) {
	case PID:
		return arg0[:], nil

	case string:
		pid, err := ParsePID(arg0)
		return pid[:], err

	default:
		t := reflect.TypeOf(arg0)
		return nil, fmt.Errorf("unsupported argument type: %v", t)
	}

}

// FromObject extracts the index value from an object. The return values
// are whether the index value was found, the index value, and any error
// while extracting the index value, respectively.
func (i PIDIndexer) FromObject(raw interface{}) (bool, []byte, error) {
	switch object := raw.(type) {
	case PID:
		return true, object[:], nil

	case string:
		index, err := base58.FastBase58Decoding(path.Base(object))
		return err == nil, index, err

	case fmt.Stringer:
		proto := object.String()
		index, err := base58.FastBase58Decoding(path.Base(proto))
		return err == nil, index, err

	case interface{ Protocol() protocol.ID }:
		proto := string(object.Protocol())
		index, err := base58.FastBase58Decoding(path.Base(proto))
		return err == nil, index, err

	default:
		t := reflect.TypeOf(raw)
		return false, nil, fmt.Errorf("unsupported object type: %v", t)
	}
}
