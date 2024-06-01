package system_test

import (
	"bytes"
	context "context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero/api"
	"github.com/wetware/go/system"
)

func TestSocket(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This is currently emulating two different things and will
	// no doubt cause confusion down the road.  Fixit whenever,
	// or not...
	deliver := &mockFn{Name: "deliver", Want: "test"}

	proc := system.SocketConfig{
		Mailbox: deliver.Mailbox(),
		Deliver: deliver,
	}.Bind(ctx)
	defer proc.Release()

	f, release := proc.Handle(ctx, func(p system.Proc_handle_Params) error {
		return p.SetEvent([]byte(deliver.Want))
	})
	defer release()
	_, err := f.Struct()
	require.NoError(t, err)

	require.Equal(t, deliver.Want, deliver.Got)
}

type mockFn struct {
	Name         string
	Buf          bytes.Buffer
	Got, Want    string
	api.Function // stub
}

func (m *mockFn) Mailbox() io.Writer {
	return &m.Buf
}

func (m *mockFn) ExportedFunction(name string) api.Function {
	if name == m.Name {
		return m
	}

	return nil
}

func (m *mockFn) Call(ctx context.Context, stack ...uint64) ([]uint64, error) {
	size := api.DecodeU32(stack[0])
	b, err := io.ReadAll(io.LimitReader(&m.Buf, int64(size)))
	if err != nil {
		return nil, err
	} else if len(b) != int(size) {
		return nil, fmt.Errorf("expected message of size %d, got %d",
			size, len(b))
	}

	m.Got = string(b)
	return nil, nil
}
