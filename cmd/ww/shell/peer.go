package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

var _ core.Invokable = (*Peer)(nil)

type Peer struct {
	Ctx  context.Context
	Host host.Host
}

// Peer methods: (peer :send "peer-addr" "proc-id" data) or (peer :connect "peer-addr")
func (p Peer) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		return p.String(), nil
	}

	if len(args) < 1 {
		return nil, fmt.Errorf("peer requires at least 1 argument: (peer :method ...)")
	}

	// First argument should be a keyword (:send, :connect, :is-self, :id)
	method, ok := args[0].(builtin.Keyword)
	if !ok {
		return nil, fmt.Errorf("peer method must be a keyword, got %T", args[0])
	}

	switch method {
	case "id":
		return p.ID(), nil
	case "send":
		if len(args) < 4 {
			return nil, fmt.Errorf("peer :send requires 3 arguments: (peer :send peer-addr proc-id data)")
		}

		var peerAddr string
		switch v := args[1].(type) {
		case string:
			peerAddr = v
		case builtin.String:
			peerAddr = string(v)
		default:
			return nil, fmt.Errorf("peer address must be a string or builtin.String, got %T", args[1])
		}

		var procID string
		switch v := args[2].(type) {
		case string:
			procID = v
		case builtin.String:
			procID = string(v)
		default:
			return nil, fmt.Errorf("process ID must be a string or builtin.String, got %T", args[2])
		}

		return p.Send(p.Ctx, peerAddr, procID, args[3])

	case "connect":
		if len(args) < 2 {
			return nil, fmt.Errorf("peer :connect requires 1 argument: (peer :connect peer-addr)")
		}

		var peerAddr string
		switch v := args[1].(type) {
		case string:
			peerAddr = v
		case builtin.String:
			peerAddr = string(v)
		default:
			return nil, fmt.Errorf("peer address must be a string or builtin.String, got %T", args[1])
		}
		return p.Connect(peerAddr)

	case "is-self":
		if len(args) < 2 {
			return nil, fmt.Errorf("peer :is-self requires 1 argument: (peer :is-self peer-id)")
		}

		var peerIDStr string
		switch v := args[1].(type) {
		case string:
			peerIDStr = v
		case builtin.String:
			peerIDStr = string(v)
		default:
			return nil, fmt.Errorf("peer ID must be a string or builtin.String, got %T", args[1])
		}
		return p.IsSelf(peerIDStr)

	default:
		return nil, fmt.Errorf("unknown peer method: %s (supported: :send, :connect, :is-self, :id)", method)
	}
}

func (p Peer) String() string {
	if p.Host == nil {
		return "<Peer: (no host)>"
	}
	return fmt.Sprintf("<Peer: %s>", p.Host.ID())
}

// Send sends data to a specific peer and process
func (p *Peer) Send(ctx context.Context, peerAddr, procIDStr string, data interface{}) (core.Any, error) {
	// Parse peer address
	peerInfo, err := p.parsePeerAddr(peerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer address: %w", err)
	}

	// Check if we're sending to ourselves
	if peerInfo.ID == p.Host.ID() {
		// TODO: Implement self-routing optimization
		// For now, we'll still go through the network
	}

	// Connect to the peer
	if err := p.Host.Connect(ctx, *peerInfo); err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %w", err)
	}

	// Create protocol ID from process ID
	protocolID := protocol.ID("/ww/0.1.0/" + procIDStr)

	// Open stream to the peer
	stream, err := p.Host.NewStream(ctx, peerInfo.ID, protocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	// Convert data to io.Reader based on type
	var reader io.Reader
	switch v := data.(type) {
	case io.Reader:
		reader = v
	case []byte:
		reader = bytes.NewReader(v)
	case string:
		reader = strings.NewReader(v)
	default:
		return nil, fmt.Errorf("unsupported data type: %T, expected io.Reader, []byte, or string", data)
	}

	// Send the data atomically
	_, err = io.Copy(stream, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to send data: %w", err)
	}

	return builtin.String("sent"), nil
}

// Connect establishes a connection to a peer
func (p *Peer) Connect(peerAddr string) (core.Any, error) {
	ctx := context.TODO()

	// Parse peer address
	peerInfo, err := p.parsePeerAddr(peerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer address: %w", err)
	}

	// Connect to the peer
	if err := p.Host.Connect(ctx, *peerInfo); err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %w", err)
	}

	return builtin.String("connected"), nil
}

// IsSelf checks if the given peer ID is our own
func (p *Peer) IsSelf(peerIDStr string) (core.Any, error) {
	targetPeerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid peer ID: %w", err)
	}

	return builtin.Bool(targetPeerID == p.Host.ID()), nil
}

// ID returns our own peer ID as a string
func (p *Peer) ID() core.Any {
	if p.Host == nil {
		return builtin.String("")
	}
	return builtin.String(p.Host.ID().String())
}

// parsePeerAddr parses a peer address (either peer ID or multiaddr) into AddrInfo
func (p *Peer) parsePeerAddr(peerAddr string) (*peer.AddrInfo, error) {
	// Try to parse as peer ID first
	peerID, err := peer.Decode(peerAddr)
	if err == nil {
		// Successfully parsed as peer ID
		return &peer.AddrInfo{
			ID: peerID,
			// Note: In a real implementation, you'd need peer discovery
			// or provide addresses as additional parameters
		}, nil
	}

	// Fall back to treating as multiaddr
	addr, err := ma.NewMultiaddr(peerAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid peer address or ID: %w", err)
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer info from multiaddr: %w", err)
	}

	return peerInfo, nil
}
