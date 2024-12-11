package system

import (
	"context"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/hashicorp/go-memdb"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/tetratelabs/wazero/api"
	"github.com/wetware/go/proc"
)

type Handler interface {
	ServeProc(context.Context, *proc.P) error
}

type Net struct {
	Host    host.Host
	Router  *memdb.MemDB
	Handler Handler
}

func (n Net) Bind(ctx context.Context, p *proc.P) (api.Closer, error) {
	tx := n.Router.Txn(true)
	defer tx.Abort()

	if err := tx.Insert("proc", p); err != nil {
		return nil, err
	} else {
		tx.Commit()
	}

	proto := Proto.Unwrap() // TODO:  ns
	n.Host.SetStreamHandlerMatch(proto,
		n.MatchProto,
		n.NewStreamHandler(ctx, n.Router))
	release := func(ctx context.Context) error {
		defer n.Host.RemoveStreamHandler(proto)

		tx := n.Router.Txn(true)
		if err := tx.Delete("proc", p); err != nil {
			tx.Abort()
			return err
		}
		tx.Commit()

		return nil
	}

	return CloserFunc(release), nil
}

func (n Net) MatchProto(id protocol.ID) bool {
	return strings.HasPrefix(string(id), Proto.String()+"/")
}

func (n Net) NewStreamHandler(ctx context.Context, db *memdb.MemDB) network.StreamHandler {
	return func(s network.Stream) {
		defer s.CloseWrite()

		log := slog.Default().With(
			"server", n.Host.ID(),
			"peer", s.Conn().RemotePeer(),
			"stream", s.ID())

		log.InfoContext(ctx, "handled stream")

		// // Apply context deadline?
		// if dl, ok := ctx.Deadline(); ok {
		// 	if err := s.SetReadDeadline(dl); err != nil {
		// 		log.WarnContext(ctx, "failed to set read deadline",
		// 			"reason", err)
		// 	}
		// }

		// proto := strings.TrimPrefix(string(s.Protocol()), Proto.String()+"/")
		// parts := strings.SplitN(proto, "/", 3) // /<index>/<id>/<method>/<stack>
		// if len(parts) < 3 {
		// 	slog.ErrorContext(ctx, "invalid protocol",
		// 		"proto", proto)
		// 	return
		// }

		// tx := db.Txn(false) // read-only
		// defer tx.Abort()

		// call, err := proc.ParseCallData(parts[3])
		// if err != nil {
		// 	slog.ErrorContext(ctx, "invalid call data",
		// 		"proto", proto,
		// 		"data", parts[3])
		// 	return
		// }

		// if v, err := tx.First("proc", parts[0], parts[1]); err != nil {
		// 	slog.ErrorContext(ctx, "failed to route message",
		// 		"reason", err,
		// 		"proto", s.Protocol())
		// } else if err = v.(*proc.P).Deliver(ctx, call, s); err != nil {
		// 	slog.ErrorContext(ctx, "failed to deliver message",
		// 		"reason", err,
		// 		"proto", s.Protocol())
		// }
	}
}

func (n Net) ServeProc(ctx context.Context, p *proc.P) error {
	if n.Handler == nil {
		return nil
	}

	return n.Handler.ServeProc(ctx, p)
}

// HostNode allows
type HostNode struct {
	Ctx     context.Context
	Host    host.Host
	Routing *memdb.MemDB
}

func (h HostNode) Open(name string) (fs.File, error) {
	if name == "." {
		panic(name) // FIXME:  probably want to return h itself here
	}

	addr, err := ma.NewMultiaddr(name)
	if err != nil {
		return nil, err
	}

	pid, err := addr.ValueForProtocol(P_PID)
	if err != nil {
		return nil, err
	}

	tx := h.Routing.Txn(false)
	defer tx.Abort()

	v, err := tx.First("proc", "pid", pid)
	if err != nil || v == nil {
		return nil, err
	}

	return ProcNode{
		P: v.(*proc.P),
	}, nil
}

type HandlerFunc func(context.Context, *proc.P) error

func (serve HandlerFunc) ServeProc(ctx context.Context, p *proc.P) error {
	return serve(ctx, p)
}
