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
	"net/url"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mr-tron/base58/base58"
	"github.com/thejerf/suture/v4"
	"github.com/wetware/go/system"
)

var httpPathPrefix = system.Proto.String() + "/proc/"

type HTTP struct {
	Env          *system.Env
	Router       Router
	ListenConfig *net.ListenConfig
	ListenAddr   string
}

func (HTTP) String() string {
	return "http"
}

func (h HTTP) Log() *slog.Logger {
	return h.Env.Log().With(
		"service", h.String(),
		"url", h.URL())
}

func (h HTTP) URL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   h.ListenAddr,
		Path:   httpPathPrefix,
	}
}

func (h HTTP) Serve(ctx context.Context) error {
	l, err := h.Listen(ctx)
	if errors.Is(err, syscall.EADDRINUSE) {
		return suture.ErrDoNotRestart // another instance is serving the namespace
	} else if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer l.Close()

	r := chi.NewRouter()
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.WriteHeader(http.StatusOK)
	})
	r.Mount(h.URL().Path, &h)

	s := &http.Server{
		Handler: r,
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
	if h.ListenConfig == nil {
		h.ListenConfig = new(net.ListenConfig)
	}
	if h.ListenAddr == "" {
		h.ListenAddr = "127.0.0.1:2080"
	}

	return h.ListenConfig.Listen(ctx, "tcp", h.ListenAddr)
}

func (h HTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	defer io.Copy(io.Discard, r.Body)

	// note trailing slash; we expect a b58 PID after it
	if !strings.HasPrefix(r.URL.Path, httpPathPrefix) {
		// 404 - Not Found
		http.Error(w, httpPathPrefix+" "+http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		// 405 - Method Not Allowed
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	pid := path.Base(r.URL.Path)
	p, err := h.Router.GetProc(pid)
	if errors.Is(err, ErrNotFound) {
		// 404 - Not Found
		http.Error(w, "Proc "+http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		// 500 - Internal Server Error
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := p.Reserve(r.Context(), r.Body); err != nil {
		// 500 - Internal Server Error
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer p.Release()

	name := r.URL.Query().Get("method")
	method := p.Method(name)
	if method == nil {
		// 400 - Bad Request
		http.Error(w, fmt.Sprintf("missing export: %s", name), http.StatusBadRequest)
		return
	}

	var stack []uint64
	for i, word := range r.URL.Query()["push"] {
		buf, err := base58.FastBase58Decoding(word)
		if err != nil {
			// 400 - Bad Request
			http.Error(w, fmt.Sprintf("stack[%d]: %v", i, err), http.StatusBadRequest)
			return
		}

		u, n := binary.Uvarint(buf)
		if n <= len(buf) {
			// 400 - Bad Request
			http.Error(w, fmt.Sprintf("stack[%d]: invalid uvarint", i), http.StatusBadRequest)
			return
		}

		stack = append(stack, u)
	}

	if err := method.CallWithStack(r.Context(), stack); err != nil {
		// 5XX - Server Error
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			// 504 - Gateway Timeout
			http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
		} else {
			// 502 - Bad Gateway
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		}
	} else {
		// 204 - No Content
		w.WriteHeader(http.StatusNoContent) // TODO:  make this 200 - OK when we're able to return results
	}
}
