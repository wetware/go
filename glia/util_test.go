package glia_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wetware/go/glia"
)

func TestFrameIO(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	frame := []byte("test data")
	err := glia.WriteFrame(buf, frame)
	require.NoError(t, err)
	require.Equal(t, len(frame)+1, buf.Len())

	got, err := glia.ReadFrame(bufio.NewReader(buf))
	require.NoError(t, err)
	require.Equal(t, frame, got)
}

func TestFrameEncoding(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected int // expected total length including uvarint
	}{
		{
			name:     "empty message",
			input:    []byte{},
			expected: 1, // just the uvarint 0
		},
		{
			name:     "small message",
			input:    []byte("hello"),
			expected: 6, // uvarint(5) + 5 bytes
		},
		{
			name:     "medium message",
			input:    bytes.Repeat([]byte("a"), 127),
			expected: 128, // uvarint(127) + 127 bytes
		},
		{
			name:     "large message",
			input:    bytes.Repeat([]byte("a"), 16384),
			expected: 16387, // uvarint(16384) = 3 bytes + 16384 bytes
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Test writing
			buf := new(bytes.Buffer)
			err := glia.WriteFrame(buf, tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, buf.Len())

			// Test reading
			got, err := glia.ReadFrame(bufio.NewReader(buf))
			require.NoError(t, err)
			require.Equal(t, tc.input, got)
		})
	}
}

func TestFrameErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("read empty buffer", func(t *testing.T) {
		buf := new(bytes.Buffer)
		_, err := glia.ReadFrame(bufio.NewReader(buf))
		require.Error(t, err)
	})

	t.Run("read truncated uvarint", func(t *testing.T) {
		buf := new(bytes.Buffer)
		// Write just the first byte of a 2-byte uvarint
		buf.WriteByte(0x80)
		_, err := glia.ReadFrame(bufio.NewReader(buf))
		require.Error(t, err)
	})

	t.Run("read truncated message", func(t *testing.T) {
		buf := new(bytes.Buffer)
		// Write uvarint for 10 bytes but only write 5
		buf.WriteByte(10)
		buf.Write(bytes.Repeat([]byte("a"), 5))
		_, err := glia.ReadFrame(bufio.NewReader(buf))
		require.Error(t, err)
	})
}
