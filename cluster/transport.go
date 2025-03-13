package cluster

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/Mathew-Estafanous/memlist"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/wetware/go/system"
	protoutils "github.com/wetware/go/util/proto"
)

var (
	_ memlist.Transport = (*HostTransport)(nil)
	_ net.Conn          = (*StreamConn)(nil)
)

var (
	packetProto = protoutils.Join(system.Proto.Unwrap(), "/memlist/packet")
	streamProto = protoutils.Join(system.Proto.Unwrap(), "/memlist/stream")
)

// StreamConn wraps a libp2p stream to implement net.Conn
type StreamConn struct {
	network.Stream
}

func (s StreamConn) LocalAddr() net.Addr {
	return HostAddr{ID: s.Conn().LocalPeer()}
}

func (s StreamConn) RemoteAddr() net.Addr {
	return HostAddr{ID: s.Conn().RemotePeer()}
}

// HostTransport implements the memlist.Transport interface using glia's process-oriented
// communication primitives. It treats each transport endpoint as a process that can
// receive packets.
type HostTransport struct {
	ctx      context.Context
	cancel   context.CancelFunc
	host     host.Host
	packetCh chan *memlist.Packet
	streamCh chan net.Conn
}

// NewHostTransport creates a new Transport implementation using glia
func NewHostTransport(ctx context.Context, h host.Host) *HostTransport {
	ctx, cancel := context.WithCancel(ctx)
	t := &HostTransport{
		host:     h,
		packetCh: make(chan *memlist.Packet, 16),
		streamCh: make(chan net.Conn, 1),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Set up protocol handlers
	t.host.SetStreamHandler(packetProto, t.HandlePacketStream)
	t.host.SetStreamHandler(streamProto, t.HandleConnStream)
	return t
}

// HandlePacketStream handles incoming packet streams by reading the packet data into a buffer
// and forwarding it to the packet channel. The stream is automatically closed when
// the handler completes. If the context is cancelled or the packet channel is full,
// the packet will be dropped.
func (t *HostTransport) HandlePacketStream(s network.Stream) {
	defer s.Close()

	// Read the packet data
	b, err := io.ReadAll(s)
	if err != nil {
		slog.Debug("failed to read packet",
			"peer", s.Conn().RemotePeer(),
			"proto", s.Protocol(),
			"reason", err)
		return
	}

	// Create and send the packet
	peer := s.Conn().RemotePeer()
	packet := &memlist.Packet{
		From: HostAddr{ID: peer},
		Buf:  b,
	}

	select {
	case t.packetCh <- packet:
	case <-t.ctx.Done():
	}
}

// HandleConnStream handles incoming stream connections by wrapping them in a StreamConn
// and forwarding them to the stream channel.
func (t *HostTransport) HandleConnStream(s network.Stream) {
	select {
	case t.streamCh <- &StreamConn{Stream: s}:
		// Stream will be closed by the receiver
	case <-t.ctx.Done():
		s.Close()
	}
}

// SendTo implements Transport.SendTo
func (t *HostTransport) SendTo(b []byte, addr string) error {
	// Parse the target address
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	// Extract the peer.ID from the multiaddr
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("failed to extract peer info: %w", err)
	}

	// Create a stream to the target process
	stream, err := t.host.NewStream(t.ctx, info.ID, packetProto)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	// Write the packet
	_, err = stream.Write(b)
	return err
}

// DialAndConnect implements Transport.DialAndConnect by establishing a stream connection
// to the specified address. The returned net.Conn can be used for bidirectional communication.
func (t *HostTransport) DialAndConnect(addr string, timeout time.Duration) (net.Conn, error) {
	// Parse the target address
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Extract the peer.ID from the multiaddr
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract peer info: %w", err)
	}

	// Create a context with timeout
	ctx := t.ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Create a stream to the target process using the stream protocol
	stream, err := t.host.NewStream(ctx, info.ID, streamProto)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return StreamConn{Stream: stream}, nil
}

// Stream implements Transport.Stream by returning a channel that receives incoming
// stream connections. Each connection can be used for bidirectional communication.
func (t *HostTransport) Stream() <-chan net.Conn {
	return t.streamCh
}

// Packets implements Transport.Packets
func (t *HostTransport) Packets() <-chan *memlist.Packet {
	return t.packetCh
}

// Shutdown implements Transport.Shutdown
func (t *HostTransport) Shutdown() error {
	// First remove the stream handlers to prevent new connections
	t.host.RemoveStreamHandler(packetProto)
	t.host.RemoveStreamHandler(streamProto)

	// Then cancel context to stop any in-flight operations
	t.cancel()

	// Finally close the channels
	close(t.packetCh)
	close(t.streamCh)
	return nil
}

// HostAddr implements net.Addr for glia process addresses
type HostAddr struct {
	peer.ID
}

func (a HostAddr) Network() string {
	return system.Proto.String()
}

func (a HostAddr) String() string {
	return a.ID.String()
}
