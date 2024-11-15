//go:generate capnp compile -I $GOPATH/src/capnproto.org/go/capnp/std -ogo proc.capnp

package proc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"path"
	"strings"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	protoutils "github.com/wetware/go/util/proto"
	"golang.org/x/sync/semaphore"
)

const DefaultMessageReadTimeout = time.Second * 30
const DefaultMaxMessageSize = 1024 << 10 // 1MB.  Arbitrary.  Can grow.

type StreamHandler struct {
	protoutils.VersionedID
	Proc               *P
	MessageReadTimeout time.Duration
	MaxMessageSize     uint32
}

func (h StreamHandler) String() string {
	parent := h.VersionedID.String()
	child := h.Proc.String()
	return path.Join(parent, child)
}

func (h StreamHandler) Proto() protocol.ID {
	pid := protocol.ID(h.Proc.Mod.Name())
	return protoutils.Join(h.Unwrap(), "proc", pid)
}

func (h StreamHandler) Match(id protocol.ID) bool {
	prefix := string(h.Proto())
	return strings.HasPrefix(string(id), prefix)
}

func (h StreamHandler) Bind(ctx context.Context) network.StreamHandler {
	if h.MessageReadTimeout <= 0 {
		h.MessageReadTimeout = DefaultMessageReadTimeout
	}

	if h.MaxMessageSize == 0 {
		h.MaxMessageSize = DefaultMaxMessageSize
	}

	var (
		// Use a weighted semaphor as a mutex because it allows asynchronous
		// acquires with ctx. FIFO message-processing is a nice side-benefit.
		mu = semaphore.NewWeighted(1)

		rd  = io.LimitedReader{N: int64(h.MaxMessageSize)}
		buf bytes.Buffer
	)
	return func(s network.Stream) {
		defer s.Close()
		defer buf.Reset()

		ctx, cancel := context.WithTimeout(ctx, h.MessageReadTimeout)
		defer cancel()

		if err := mu.Acquire(ctx, 1); err != nil {
			slog.DebugContext(ctx, "closing stream",
				"reason", err,
				"stream", s.ID())
			return
		}
		defer mu.Release(1)

		d := time.Now().Add(h.MessageReadTimeout)
		if err := s.SetReadDeadline(d); err != nil {
			slog.ErrorContext(ctx, "failed to set delivery deadline",
				"reason", err,
				"stream", s.ID())
			return
		}

		// Read the message into a local buffer
		n, err := io.Copy(&buf, &rd)
		if err != nil {
			slog.ErrorContext(ctx, "failed to read message",
				"reason", err,
				"stream", s.ID(),
				"n_bytes", n)
			return
		} else if n > (1<<32 - 1) { // max uint32
			slog.ErrorContext(ctx, "failed to read message",
				"reason", errors.New("size overflows u32"),
				"stream", s.ID(),
				"n_bytes", n)
			return
		}

		// Unmarshal the method call from the buffer
		m, err := capnp.Unmarshal(buf.Bytes())
		if err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal capnp message",
				"reason", err,
				"stream", s.ID())
			return
		}
		defer m.Release()

		call, err := ReadRootMethodCall(m)
		if err != nil {
			slog.ErrorContext(ctx, "failed to read root method call",
				"reason", err,
				"stream", s.ID())
			return
		}

		// copy the stream to the process' mailbox
		err = h.Proc.Deliver(ctx, call)
		if errors.Is(err, context.DeadlineExceeded) {
			slog.DebugContext(ctx, "closing stream",
				"reason", err,
				"stream", s.ID())
		} else if err != nil {
			slog.ErrorContext(ctx, "message delivery failed",
				"reason", err,
				"stream", s.ID())
		}
	}
}
