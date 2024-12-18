package glia

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"strings"

	capnp "capnproto.org/go/capnp/v3"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/wetware/go/system"
)

var ErrNotFound = errors.New("not found")
var ErrU16Overflow = errors.New("varint overflows u16")

type RPC struct {
	Env    *system.Env
	Router Router
}

func (rpc RPC) Log() *slog.Logger {
	return rpc.Env.Log().With(
		"service", rpc.String())
}

func (rpc RPC) String() string {
	return "rpc"
}

func (rpc RPC) Protocol() protocol.ID {
	return system.Proto.Unwrap()
}

func (rpc RPC) Match(id protocol.ID) bool {
	prefix := system.Proto.String()
	return strings.HasPrefix(string(id), prefix)
}

func (rpc RPC) Serve(ctx context.Context) error {
	proto := rpc.Protocol()
	rpc.Env.Host.SetStreamHandlerMatch(proto, rpc.Match, func(s network.Stream) {
		defer s.Close()

		if dl, ok := ctx.Deadline(); ok {
			if err := s.SetDeadline(dl); err != nil {
				rpc.Log().WarnContext(ctx, "failed to set deadline",
					"reason", err)
				// non-fatal; continue along...
			}
		}

		if err := rpc.ServeStream(ctx, s); err != nil {
			rpc.Log().ErrorContext(ctx, "failed to serve stream",
				"reason", err,
				"stream", s.ID())
		}
	})
	defer rpc.Env.Host.RemoveStreamHandler(proto)
	rpc.Log().DebugContext(ctx, "service started")

	<-ctx.Done()
	return nil

	// children := suture.New(rpc.String(), suture.Spec{
	// 	EventHook: util.EventHookWithContext(ctx),
	// })
	// children.Add(&HTTPServer{Env: rpc.Env})
	// children.Add(&UnixServer{Env: rpc.Env})

	// return children.Serve(ctx)
}

func (rpc RPC) ServeStream(ctx context.Context, s network.Stream) error {
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
	return rpc.ServeP2P(w, req)
}

func (rpc RPC) ServeP2P(w io.WriteCloser, req *Request) error {
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

	if r := rpc.Bind(req); r != nil {
		if err := r.Render(req.Ctx, res); err != nil {
			return err
		}
	}

	b, err := res.Message().Marshal()
	if err != nil {
		return err
	}

	_, err = io.Copy(w, bytes.NewReader(b))
	return err
}

func (rpc RPC) Bind(req *Request) Renderer {
	p, err := rpc.Router.GetProc(req.PID)
	if err != nil {
		return RoutingError(err)
	}

	name, err := req.Call.Method()
	if err != nil {
		return InvalidMethod(err)
	}

	mc := MethodCall{
		P:      p,
		Method: name,
		Body:   &req.Body,
	}

	stack, err := req.Call.Stack()
	if err != nil {
		return InvalidCallStack(err)
	}
	for i := 0; i < stack.Len(); i++ {
		mc.Stack = append(mc.Stack, stack.At(i))
	}

	return mc
}

type Request struct {
	Ctx  context.Context
	PID  string
	Call CallData
	Body bufio.Reader
}

func ReadRequest(ctx context.Context, s network.Stream) (*Request, error) {
	pid := path.Base(string(s.Protocol()))
	req := &Request{Ctx: ctx, PID: pid}
	req.Body.Reset(s)

	m, err := ReadMessage(&req.Body)
	if err != nil {
		return nil, err
	}

	req.Call, err = ReadRootCallData(m)
	return req, err
}

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
