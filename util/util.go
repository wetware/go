package util

import (
	"encoding/binary"
	"fmt"
)

func Uint64SliceToByteSlice(u []uint64) []byte {
	bytes := make([]byte, len(u)*8)
	for i, v := range u {
		binary.LittleEndian.PutUint64(bytes[i*8:], v)
	}
	return bytes
}

func ByteSliceToUint64Slice(b []byte) ([]uint64, error) {
	if len(b)%8 != 0 {
		return nil, fmt.Errorf("not a multiple of 8 %d", len(b))
	}

	uint64s := make([]uint64, len(b)/8)
	for i := range uint64s {
		uint64s[i] = binary.LittleEndian.Uint64(b[i*8 : (i+1)*8])
	}
	return uint64s, nil
}
