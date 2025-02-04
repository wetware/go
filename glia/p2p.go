package glia

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	capnp "capnproto.org/go/capnp/v3"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/wetware/go/system"
)

var ErrU16Overflow = errors.New("varint overflows u16")

type P2P struct {
	Env    *system.Env
	Router Router
}

func (p2p P2P) Log() *slog.Logger {
	return p2p.Env.Log().With(
		"service", p2p.String())
}

func (p2p P2P) String() string {
	return "p2p"
}

func (p2p P2P) Match(id protocol.ID) bool {
	prefix := system.Proto.String()
	return strings.HasPrefix(string(id), prefix)
}

func (p2p P2P) Serve(ctx context.Context) error {
	proto := system.Proto.Unwrap()
	p2p.Env.Host.SetStreamHandler(proto, func(s network.Stream) {
		defer s.Close()

		if dl, ok := ctx.Deadline(); ok {
			if err := s.SetDeadline(dl); err != nil {
				p2p.Log().WarnContext(ctx, "failed to set deadline",
					"reason", err)
				// non-fatal; continue along...
			}
		}

		if err := p2p.ServeStream(ctx, s); err != nil {
			p2p.Log().ErrorContext(ctx, "failed to serve stream",
				"reason", err,
				"stream", s.ID())
		}
	})
	defer p2p.Env.Host.RemoveStreamHandler(proto)
	p2p.Log().DebugContext(ctx, "service started")

	<-ctx.Done()
	return nil

	// children := suture.New(p2p.String(), suture.Spec{
	// 	EventHook: util.EventHookWithContext(ctx),
	// })
	// children.Add(&HTTPServer{Env: p2p.Env})
	// children.Add(&UnixServer{Env: p2p.Env})

	// return children.Serve(ctx)
}

func (p2p P2P) ServeStream(ctx context.Context, s network.Stream) error {
	// Glia RPC is a synchronous RPC protocol models one round-trip
	// (request-response) between a server and a client.  The round-
	// trip models a synchronous method call on an object.
	////

	// 1. Declare a response writer, which is responsible for writing
	//    a response back to the client.
	////
	w := &ResponseWriter{Stream: s}
	defer w.Close()

	// 2. Read the request headers from the incoming stream.
	////
	req, err := ReadRequest(ctx, s)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer req.Body.Reset(nil)

	// 3. Serve the request.
	////
	return p2p.ServeP2P(w, req)
}

func (p2p P2P) ServeP2P(w io.Writer, req *Request) error {
	// We always return a result of some kind.  The first task
	// is to set up a response arena for the request. All data
	// written to this arena is ultimately sent to the remote
	// caller.
	////

	m, s := capnp.NewSingleSegmentMessage(nil)
	defer m.Release()

	res, err := NewRootResult(s)
	if err != nil {
		return err
	}

	// Render the method call
	////
	method, err := p2p.Bind(req)
	if err != nil {
		return err
	}
	if err := method.Render(req.Ctx, res); err != nil {
		return err
	}

	// Return results
	////
	b, err := res.Message().Marshal()
	if err != nil {
		return err
	}

	// wrap b in a frame consisiting of a uvarint length prefix.
	frame := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(frame, uint64(len(b)))
	frame = append(frame[:n], b...)

	// write frame to w.
	if _, err = io.Copy(w, bytes.NewReader(frame)); err == nil { // write results
		_, err = io.Copy(w, method.P.OutBuffer()) // write body
	}

	return err
}

func (p2p P2P) Bind(req *Request) (*MethodCall, error) {
	pid, err := req.Header.Proc()
	if err != nil {
		return nil, err
	}

	p, err := p2p.Router.GetProc(pid)
	if err != nil {
		return nil, err
	}

	name, err := req.Header.Method()
	if err != nil {
		return nil, err
	}

	mc := &MethodCall{
		P:      p,
		Method: name,
		Body:   &req.Body,
	}

	stack, err := req.Header.Stack()
	if err == nil {
		for i := 0; i < stack.Len(); i++ {
			mc.Stack = append(mc.Stack, stack.At(i))
		}
	}
	return mc, err
}

type Request struct {
	Ctx    context.Context
	Header Header
	Body   bufio.Reader
}

func ReadRequest(ctx context.Context, s network.Stream) (*Request, error) {
	req := &Request{Ctx: ctx}
	req.Body.Reset(s)

	m, err := ReadMessage(&req.Body)
	if err != nil {
		return nil, err
	}

	req.Header, err = ReadRootHeader(m)
	return req, err
}

// func (req Request) Render(ctx context.Context, res Result) error {

// }

type ResponseWriter struct {
	Stream interface {
		io.Writer
		CloseWrite() error
	}
}

func (rw ResponseWriter) Close() error {
	return rw.Stream.CloseWrite()
}

func (rw ResponseWriter) Write(p []byte) (int, error) {
	return rw.Stream.Write(p)
}
