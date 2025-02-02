package glia

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	capnp "capnproto.org/go/capnp/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/mr-tron/base58"
	"github.com/wetware/go/system"
)

const DefaultListenAddr = "ww.local:8020"

var ErrNotFound = errors.New("not found")

type HTTP struct {
	P2P P2P

	once         sync.Once
	ListenConfig *net.ListenConfig
	ListenAddr   string
	Router       chi.Router
}

func (*HTTP) String() string {
	return "http"
}

func (h *HTTP) Log() *slog.Logger {
	return h.P2P.Env.Log().With(
		"service", h.String(),
		"addr", h.ListenAddr)
}

func (h *HTTP) Init() {
	h.once.Do(func() {
		if h.ListenConfig == nil {
			h.ListenConfig = &net.ListenConfig{}
		}

		if h.ListenAddr == "" {
			h.ListenAddr = DefaultListenAddr
		}

		if h.Router == nil {
			h.Router = h.DefaultRouter()
		}
	})
}

func (h *HTTP) DefaultRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.WriteHeader(http.StatusNoContent)
	})
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		protoVersion := system.Proto.String()
		if n, err := io.Copy(w, strings.NewReader(protoVersion)); err == nil {
			h.Log().Debug("failed to write response",
				"endpoint", "/version",
				"wrote", n,
				"reason", err)
		}
	})

	path := path.Join("/", system.Proto.String(), ":host/:proc/:method")
	r.Post(path, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		defer io.Copy(io.Discard, r.Body)

		// Bind request data
		////
		m, seg := capnp.NewSingleSegmentMessage(nil)
		defer m.Release()

		req, err := NewMessageRoutingRequest(r, seg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := render.Bind(r, &req); err != nil {
			if errors.Is(err, ErrNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Render the request
		////
		if err := h.P2P.ServeP2P(w, &req.GliaRequest); err != nil {
			if errors.Is(err, ErrNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return r
}

func (h *HTTP) Serve(ctx context.Context) error {
	l, err := h.Listen(ctx)
	if errors.Is(err, syscall.EADDRINUSE) {
		// Another wetware instance is serving the HTTP API on
		// this address.
		//
		// Let it Fail (TM)
		//
		// The service is restarted with exponential backoff, so
		// if the other process fails, this one will eventually
		// take over.
		//
		// This isn't a failure, so log the event and swallow the
		// error.
		////
		h.Log().Info("disabled HTTP service",
			"reason", err)
		return nil

	} else if err != nil {
		// Something actually went wrong and we need to fail
		// the service, e.g. a context cancellation.
		////
		return fmt.Errorf("listen: %w", err)
	}
	defer l.Close()

	s := &http.Server{
		Handler: h.Router,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// fork a goroutine to babysit the context and shut down the
	// server when it expires.  Technically there is a race condition
	// wherein s.Shutdown() can be called before s.Serve(), but I've
	// not observed it in practice, and the listener is going to be
	// forcibly closed when Serve returns, anyway.
	////
	cherr := make(chan error, 1)
	go func() {
		defer close(cherr)
		<-ctx.Done()

		// Supervisor defaults to 10s shutdown timeout.  Use half of
		// that.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		cherr <- s.Shutdown(ctx)
	}()
	h.Log().DebugContext(ctx, "service started")

	// Serve over the listener.  This is a blocking call
	// that always returns a non-nil error.
	if err := s.Serve(l); err != http.ErrServerClosed {
		return err
	}
	cancel()

	// If we get here, then s.Shutdown() was called, and we're waiting
	// to see if it returned an error.
	return <-cherr
}

func (h *HTTP) Listen(ctx context.Context) (net.Listener, error) {
	h.Init()
	return h.ListenConfig.Listen(ctx, "tcp", h.ListenAddr)
}

type MessageRoutingRequest struct {
	GliaRequest Request
}

func NewMessageRoutingRequest(r *http.Request, seg *capnp.Segment) (MessageRoutingRequest, error) {
	hdr, err := NewRootHeader(seg)
	return MessageRoutingRequest{GliaRequest: Request{Header: hdr}}, err
}

func (req *MessageRoutingRequest) Bind(r *http.Request) error {
	if err := req.GliaRequest.Header.SetProc(chi.URLParam(r, "proc")); err != nil {
		return fmt.Errorf("set proc: %w", err)
	}
	if err := req.GliaRequest.Header.SetMethod(chi.URLParam(r, "method")); err != nil {
		return fmt.Errorf("set method: %w", err)
	}

	pushVals := r.URL.Query()["push"]
	stackSize := int32(len(pushVals))
	stack, err := req.GliaRequest.Header.NewStack(stackSize)
	if err != nil {
		return fmt.Errorf("new stack: %w", err)
	}
	for i, word := range pushVals {
		buf, err := base58.FastBase58Decoding(word)
		if err != nil {
			return fmt.Errorf("stack[%d]: %w", i, err)
		}

		u, n := binary.Uvarint(buf)
		if n <= 0 {
			return fmt.Errorf("stack[%d]: invalid uvarint", i)
		}

		stack.Set(i, u)
	}

	req.GliaRequest.Ctx = r.Context()
	req.GliaRequest.Body.Reset(r.Body)
	return nil
}
