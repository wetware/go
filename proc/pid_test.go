package proc_test

import (
	"bytes"
	"testing"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/proc"
)

func TestPID(t *testing.T) {
	t.Parallel()
	t.Helper()

	want := proc.NewPID()

	t.Run("String", func(t *testing.T) {
		b, err := base58.FastBase58Decoding(want.String())
		require.NoError(t, err)
		require.Equal(t, want[:], b)
	})

	t.Run("Read", func(t *testing.T) {
		r := bytes.NewReader(want[:])
		got, err := proc.ReadPID(r)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("Parse", func(t *testing.T) {
		got, err := proc.ParsePID(want.String())
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

func TestPIDIndexer(t *testing.T) {
	t.Parallel()
	t.Helper()

	pid := proc.NewPID()
	ix := proc.PIDIndexer{}

	t.Run("FromArgs", func(t *testing.T) {
		index, err := ix.FromArgs(pid)
		require.NoError(t, err)
		require.Equal(t, pid[:], index)
		require.Len(t, index, 20, "PID should be 160 bits long")

		index, err = ix.FromArgs(pid.String())
		require.NoError(t, err)
		require.Equal(t, pid[:], index)
		require.Len(t, index, 20, "PID should be 160 bits long")

		// Failure modes
		////

		// too many arguments
		index, err = ix.FromArgs(pid, "unexpected", "extra", "args")
		require.Nil(t, index)
		require.Error(t, err)

		// missing argument
		index, err = ix.FromArgs()
		require.Nil(t, index)
		require.Error(t, err)

		// unsupported arg type
		index, err = ix.FromArgs(struct{}{})
		require.Nil(t, index)
		require.Error(t, err)
	})

	t.Run("FromArgs", func(t *testing.T) {
		ok, index, err := ix.FromObject(pid)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, pid[:], index)
		require.Len(t, index, 20, "PID should be 160 bits long")

		ok, index, err = ix.FromObject(pid.String())
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, pid[:], index)
		require.Len(t, index, 20, "PID should be 160 bits long")

		rd := bytes.NewBufferString(pid.String())
		ok, index, err = ix.FromObject(rd)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, pid[:], index)
		require.Len(t, index, 20, "PID should be 160 bits long")

		id := mockProto(pid.String())
		ok, index, err = ix.FromObject(id)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, pid[:], index)
		require.Len(t, index, 20, "PID should be 160 bits long")

		// Failure modes
		////

		// unsupported object type
		ok, index, err = ix.FromObject(struct{}{})
		require.Nil(t, index)
		require.False(t, ok)
		require.Error(t, err)
	})
}

type mockProto protocol.ID

func (p mockProto) Protocol() protocol.ID {
	return protocol.ID(p)
}
