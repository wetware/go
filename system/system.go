//go:generate capnp compile -I$GOPATH/src/capnproto.org/go/capnp/std -ogo system.capnp

package system

import (
	"bytes"
	"context"
	"io"
	"log/slog"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/exp/bufferpool"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const ModuleName = "ww"

type Module struct {
	api.Module
	stdin bytes.Buffer
}

func (m *Module) Stdin() io.Reader {
	return &m.stdin
}

func (m *Module) Mailbox() io.Writer {
	return &m.stdin
}

func (m *Module) SocketConfig(mod api.Module) SocketConfig {
	return SocketConfig{
		System: m,
		Guest:  mod,
	}
}

// Boot returns the system's bootstrap client.  This capability is
// analogous to the "root" user in a Unix system.
func (m *Module) Boot(mod api.Module) capnp.Client {
	socket := m.SocketConfig(mod).Build()
	server := Proc_NewServer(socket)
	return capnp.NewClient(server)
}

type Builder struct {
	// IPFS iface.CoreAPI
	// Host host.Host
}

func (b Builder) Compile(ctx context.Context, r wazero.Runtime) (wazero.CompiledModule, error) {
	return b.HostModuleBuilder(r).Compile(ctx)
}

func (b Builder) Instantiate(ctx context.Context, r wazero.Runtime) (*Module, error) {
	mod, err := b.HostModuleBuilder(r).Instantiate(ctx)
	return &Module{Module: mod}, err
}

func (b Builder) HostModuleBuilder(r wazero.Runtime) wazero.HostModuleBuilder {
	hmb := r.NewHostModuleBuilder(ModuleName)
	return b.WithExports(hmb)
}

func (b Builder) WithExports(hmb wazero.HostModuleBuilder) wazero.HostModuleBuilder {
	hmb = b.WithTransport(hmb)
	return hmb
}

func (b Builder) WithTransport(hmb wazero.HostModuleBuilder) wazero.HostModuleBuilder {
	return WithExports(hmb, send)
}

func WithExports(hmb wazero.HostModuleBuilder, exports ...*HostFunc) wazero.HostModuleBuilder {
	for _, e := range exports {
		hmb = hmb.NewFunctionBuilder().
			WithName(e.Name).
			WithResultNames(e.ResultNames...).
			WithParameterNames(e.ParamNames...).
			WithGoModuleFunction(e.Fn, e.ParamTypes, e.ResultTypes).
			Export(e.Name)
	}

	return hmb
}

var send = &HostFunc{
	Name:        "send",
	ParamNames:  []string{"offset", "length"},
	ParamTypes:  []api.ValueType{api.ValueTypeI32, api.ValueTypeI32},
	ResultNames: nil,
	ResultTypes: nil,
	Fn: func(ctx context.Context, mod api.Module, stack []uint64) {
		offset := api.DecodeU32(stack[0])
		length := api.DecodeU32(stack[1])

		// NOTE:  b is safe to use until the next time the guest
		// executes.  Beyond this point, guest runtimes may have
		// re-allocated the bytes indexed by b.
		b, ok := mod.Memory().Read(offset, length)
		if !ok {
			slog.ErrorContext(ctx, "out-of-bounds memory access",
				"offset", offset,
				"length", length)
			return
		}

		buf := bufferpool.Default.Get(len(b))
		buf = buf[:copy(buf, b)] // defensive indexing

		if err := SendMail(ctx, buf); err != nil {
			slog.DebugContext(ctx, "send failed",
				"offset", offset,
				"length", length)
			return
		}
	},
}

func WithMailbox(ctx context.Context, ch chan []byte) context.Context {
	return context.WithValue(ctx, keyMailbox{}, ch)
}

func SendMail(ctx context.Context, b []byte) error {
	recver := ctx.Value(keyMailbox{}).(chan []byte)
	select {
	case recver <- b:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type keyMailbox struct{}

type HostFunc struct {
	Name        string
	ParamNames  []string
	ParamTypes  []api.ValueType
	ResultNames []string
	ResultTypes []api.ValueType
	Fn          api.GoModuleFunc
}
