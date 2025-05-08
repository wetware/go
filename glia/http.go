package glia

import (
	"context"
	"encoding/json"
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

	"github.com/go-chi/chi/v5"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/wetware/go/system"
	"go.uber.org/multierr"
)

const DefaultListenAddr = "ww.local:8020"

var ErrNotFound = errors.New("not found")

type HTTP struct {
	Env    Env
	Router system.Router
	Root   string

	once         sync.Once
	ListenConfig *net.ListenConfig
	ListenAddr   string
	Handler      http.Handler
}

func (*HTTP) String() string {
	return "http"
}

func (h *HTTP) Log() *slog.Logger {
	p2p := P2P{Env: h.Env}
	return p2p.Env.Log().With(
		"service", h.String(),
		"addr", h.ListenAddr)
}

func (h *HTTP) Listen(ctx context.Context) (net.Listener, error) {
	h.Init()
	return h.ListenConfig.Listen(ctx, "tcp", h.ListenAddr)
}

func (h *HTTP) Init() {
	h.once.Do(func() {
		if h.ListenConfig == nil {
			h.ListenConfig = &net.ListenConfig{}
		}

		if h.ListenAddr == "" {
			h.ListenAddr = DefaultListenAddr
		}

		if h.Handler == nil {
			h.Handler = h.DefaultRouter()
		}
	})
}

func (h *HTTP) DefaultRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/status", h.status)
	r.Get("/info", h.info)
	r.Get("/version", h.version)
	r.Get("/root", h.root)

	path := path.Join(system.Proto.Path(), "{host}/{proc}/{method}")
	r.Post(path, h.glia)

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
		Handler: h.Handler,
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

// status handles GET requests to the /status endpoint.
// It returns a 204 No Content response to indicate the service is running.
// This endpoint is useful for health checks and service availability monitoring.
func (h *HTTP) status(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// info handles GET requests to the /info endpoint.
// It returns a 200 OK response with a JSON object containing the host's peer information,
// including its ID and network addresses. The response is encoded as a peer.AddrInfo object.
// This endpoint is useful for discovering the host's identity and network configuration.
func (h *HTTP) info(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// Get host info
	info := host.InfoFromHost(h.Env.LocalHost())

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and write the response
	if err := json.NewEncoder(w).Encode(info); err != nil {
		h.Log().ErrorContext(r.Context(), "failed to write info response",
			"reason", err)
	}
}

func (h *HTTP) version(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	protoVersion := system.Proto.String()
	if n, err := io.Copy(w, strings.NewReader(protoVersion)); err == nil {
		h.Log().Debug("failed to write response",
			"endpoint", "/version",
			"wrote", n,
			"reason", err)
	}
}

func (h *HTTP) glia(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	defer io.Copy(io.Discard, r.Body)

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	p2p := P2P{Env: h.Env, Router: h.Router}
	if err := p2p.ServeStream(r.Context(), HTTPStream{
		ResponseWriter: w,
		Request:        r,
	}); err != nil {
		h.Log().ErrorContext(r.Context(), "stream handler failed",
			"reason", err)
		return
	}
}

type HTTPStream struct {
	http.ResponseWriter
	*http.Request
}

var _ Stream = (*HTTPStream)(nil)

func (s HTTPStream) Protocol() protocol.ID {
	path := s.Request.URL.Path
	return protocol.ID(path)
}

func (s HTTPStream) Destination() string {
	hostID := s.Request.PathValue("host")
	return hostID
}

func (s HTTPStream) ProcID() string {
	return s.Request.PathValue("proc")
}

func (s HTTPStream) MethodName() string {
	return s.Request.PathValue("method")
}

func (s HTTPStream) Close() error {
	return multierr.Combine(
		s.CloseRead(),
		s.CloseWrite())
}

func (s HTTPStream) CloseRead() error {
	return s.Request.Body.Close()
}

func (s HTTPStream) CloseWrite() error {
	if c, ok := s.ResponseWriter.(io.Closer); ok {
		return c.Close()
	}

	// TODO:  might need to hook in here to signal clean exit
	return nil
}

func (s HTTPStream) Read(p []byte) (int, error) {
	return s.Request.Body.Read(p)
}

func (s HTTPStream) Write(p []byte) (int, error) {
	return s.ResponseWriter.Write(p)
}

// root handles GET requests to the /root endpoint.
// It returns a 200 OK response with the root proc.ID.
func (h *HTTP) root(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Write the root proc.ID as a base58-encoded string
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, h.Root); err != nil {
		h.Log().ErrorContext(r.Context(), "failed to write root response",
			"reason", err)
	}
}
