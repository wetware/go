package vat

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/tetratelabs/wazero/api"
	"github.com/wetware/go/system"
)

const Proto = "/ww/0.0.0"

var _ rpc.Network = (*Network)(nil)

type NetConfig struct {
	Host        host.Host
	Guest       api.Module
	System      *system.Module
	DialTimeout time.Duration
}

func (c NetConfig) Proto() protocol.ID {
	return ProtoFromModule(c.Guest)
}

func (c NetConfig) Build(ctx context.Context) Network {
	if c.DialTimeout <= 0 {
		c.DialTimeout = time.Second * 10
	}

	l := ListenConfig{
		Host: c.Host,
	}.Listen(ctx, c.Proto())

	return Network{
		NetConfig: c,
		Listener:  l,
	}
}

type Network struct {
	Host host.Host
	NetConfig
	Listener
}

func (n Network) String() string {
	return fmt.Sprintf("Vat{%s}", n.Host.ID())
}

// Return the identifier for caller on this network.
func (n Network) LocalID() rpc.PeerID {
	return rpc.PeerID{Value: n.Host.ID()}
}

func (n Network) BootstrapClient() capnp.Client {
	proc := n.System.Bind(n.Guest)
	return capnp.Client(proc)
}

func (n Network) Serve(ctx context.Context) error {
	proc := n.System.Bind(n.Guest)
	defer proc.Release()

	for c := capnp.Client(proc); ; {
		if conn, err := n.Accept(ctx, &rpc.Options{
			BootstrapClient: c.AddRef(),
			Network:         n,
		}); err == nil {
			go n.ServeConn(ctx, conn)
		} else {
			return err
		}
	}
}

func (n Network) ServeConn(ctx context.Context, conn *rpc.Conn) {
	defer conn.Close()
	slog.InfoContext(ctx, "accepted",
		"from", conn.RemotePeerID().Value)

	select {
	case <-ctx.Done():
		slog.DebugContext(ctx, "hung up on remote peer",
			"reason", ctx.Err())

	case <-conn.Done():
		slog.DebugContext(ctx, "remote peer hung up",
			"reason", conn.Close())
	}
}

// Bind default values to opt.  If opt == nil, default values
// produced by v.Options() are used.  The opt.Network and the
// opt.RemotePeerID fields are always overridden.
func (n Network) Bind(opt *rpc.Options, remote rpc.PeerID) {
	if opt == nil {
		opt = &rpc.Options{}
	}
	opt.Network = n
	opt.RemotePeerID = remote
}

// Connect to another peer by ID. The supplied Options are used
// for the connection, with the values for RemotePeerID and Network
// overridden by the Network.
func (n Network) Dial(id rpc.PeerID, opt *rpc.Options) (*rpc.Conn, error) {
	n.Bind(opt, id)

	ctx, cancel := context.WithTimeout(context.Background(), n.DialTimeout)
	defer cancel()

	pid := id.Value.(peer.ID)

	stream, err := n.Host.NewStream(ctx, pid, n.Proto())
	if err != nil {
		return nil, err
	}

	return rpc.NewConn(rpc.NewPackedStreamTransport(stream), opt), nil
}

// Accept the next incoming connection on the network, using the
// supplied Options for the connection. Generally, callers will
// want to invoke this in a loop when launching a server.
func (n Network) Accept(ctx context.Context, opt *rpc.Options) (*rpc.Conn, error) {
	s, err := n.Listener.Accept(ctx)
	if err != nil {
		return nil, err
	}

	id := s.Conn().RemotePeer()
	n.Bind(opt, rpc.PeerID{Value: id})
	conn := rpc.NewConn(rpc.NewPackedStreamTransport(s), opt)

	return conn, nil
}

// Introduce the two connections, in preparation for a third party
// handoff. Afterwards, a Provide messsage should be sent to
// provider, and a ThirdPartyCapId should be sent to recipient.
func (n Network) Introduce(provider, recipient *rpc.Conn) (rpc.IntroductionInfo, error) {
	panic("NOT IMPLEMENTED") // TODO
}

// Given a ThirdPartyCapID, received from introducedBy, connect
// to the third party. The caller should then send an Accept
// message over the returned Connection.
func (n Network) DialIntroduced(capID rpc.ThirdPartyCapID, introducedBy *rpc.Conn) (*rpc.Conn, rpc.ProvisionID, error) {
	panic("NOT IMPLEMENTED") // TODO
}

// Given a RecipientID received in a Provide message via
// introducedBy, wait for the recipient to connect, and
// return the connection formed. If there is already an
// established connection to the relevant Peer, this
// SHOULD return the existing connection immediately.
func (n Network) AcceptIntroduced(recipientID rpc.RecipientID, introducedBy *rpc.Conn) (*rpc.Conn, error) {
	panic("NOT IMPLEMENTED") // TODO
}
