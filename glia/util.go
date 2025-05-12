package glia

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
)

func ReadFrame(r *bufio.Reader) ([]byte, error) {
	size, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, size)

	_, err = io.ReadFull(r, buf)
	return buf, err
}

func WriteFrame(w io.Writer, body []byte) error {
	// Write uvarint header
	////
	var buf [binary.MaxVarintLen64]byte
	size := uint64(len(body))
	n := binary.PutUvarint(buf[:], size)
	if _, err := io.Copy(w, bytes.NewReader(buf[:n])); err != nil {
		return err
	}

	// Write body
	////
	_, err := io.Copy(w, bytes.NewReader(body))
	return err
}
