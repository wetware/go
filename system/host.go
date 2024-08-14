package system

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type HostConfig struct {
	NS   string
	Host host.Host
}

func (c HostConfig) Instantiate(ctx context.Context, r wazero.Runtime) (api.Module, error) {
	mod, err := r.NewHostModuleBuilder("ww").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(c.Send),
			[]api.ValueType{MemorySegment(0).ValueType()}, // params
			[]api.ValueType{}). // return values
		Export("send").
		Instantiate(ctx)
	return &module{mod}, err
}

func (c HostConfig) Send(ctx context.Context, mod api.Module, stack []uint64) {
	if len(stack) != MemorySegment(0).NumWords() {
		slog.ErrorContext(ctx, "unable to send message",
			"reason", "unexpected number of parameters to host function ww::send",
			"stack", stack)
		return
	}

	mem := mod.Memory()
	seg := MemorySegment(stack[0])
	msg, ok := seg.Load(mem)
	if !ok {
		slog.ErrorContext(ctx, "out-of-bounds memory access",
			"offset", seg.Offset(),
			"length", seg.Length())
		return
	}

	fmt.Println(string(msg))

	// TODO:  gotta send the message somewhere

	// s, err := c.Host.NewStream(ctx, message.To(), message.Protos()...) // TODO:  do we need to cache streams?
	// if err != nil {
	// 	// mod.CloseWithExitCode()
	// }
}

type MemorySegment uint64

func NewMemorySegment(offset, length uint32) MemorySegment {
	s := MemorySegment(offset) << 32
	s |= MemorySegment(length)
	return s
}

func (MemorySegment) NumWords() int {
	return 1
}

func (MemorySegment) ValueType() api.ValueType {
	return api.ValueTypeI64
}

func (s MemorySegment) Offset() uint32 {
	return uint32(s >> 32) // 32 leftmost bits
}

func (s MemorySegment) Length() uint32 {
	return uint32(s) // 32 rightmost bits
}

func (s MemorySegment) Load(mem api.Memory) ([]byte, bool) {
	offset := s.Offset()
	length := s.Length()
	return mem.Read(offset, length)
}

type module struct {
	api.Module
}

func (m module) Close(ctx context.Context) error {
	return m.Module.Close(ctx)
}
