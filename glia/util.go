package glia

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	capnp "capnproto.org/go/capnp/v3"
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

func ReadMessage(r *bufio.Reader) (*capnp.Message, error) {
	body, err := ReadFrame(r)
	if err != nil {
		return nil, err
	}

	return capnp.Unmarshal(body)
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

func WriteMessage(w io.Writer, m *capnp.Message) error {
	b, err := m.Marshal()
	if err != nil {
		return err
	}

	return WriteFrame(w, b)
}
