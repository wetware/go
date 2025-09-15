package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spy16/slurp"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

const helpMessage = `Wetware Shell - Available commands:
help                    - Show this help message
version                 - Show wetware version
(+ a b ...)            - Sum numbers
(* a b ...)            - Multiply numbers
(= a b)                - Compare equality
(> a b)                - Greater than
(< a b)                - Less than
(println expr)         - Print expression with newline
(print expr)           - Print expression without newline
(send "peer-addr-or-id" "proc-id" data) - Send data to a peer process (data: string, []byte, or io.Reader)
(import "module")      - Import a module (stubbed)

IPFS Path Syntax:
/ipfs/QmHash/...       - Direct IPFS path
/ipns/domain/...       - IPNS path`

var globals = map[string]core.Any{
	// Basic values
	"nil":     builtin.Nil{},
	"true":    builtin.Bool(true),
	"false":   builtin.Bool(false),
	"version": builtin.String("wetware-0.1.0"),

	// Basic operations
	"=": slurp.Func("=", core.Eq),
	"+": slurp.Func("sum", func(a ...int) int {
		sum := 0
		for _, item := range a {
			sum += item
		}
		return sum
	}),
	">": slurp.Func(">", func(a, b builtin.Int64) bool {
		return a > b
	}),
	"<": slurp.Func("<", func(a, b builtin.Int64) bool {
		return a < b
	}),
	"*": slurp.Func("*", func(a ...int) int {
		product := 1
		for _, item := range a {
			product *= item
		}
		return product
	}),
	"/": slurp.Func("/", func(a, b builtin.Int64) float64 {
		return float64(a) / float64(b)
	}),

	// Wetware-specific functions
	"help": slurp.Func("help", func() string {
		return helpMessage
	}),
	"println": slurp.Func("println", func(args ...core.Any) {
		for _, arg := range args {
			fmt.Println(arg)
		}
	}),
	"print": slurp.Func("print", func(args ...core.Any) {
		for _, arg := range args {
			fmt.Print(arg)
		}
	}),
	"send": slurp.Func("send", func(peerAddr, procId string, data interface{}) error {
		return SendToPeer(peerAddr, procId, data)
	}),
}

// SendToPeer sends data to a specific peer and process
func SendToPeer(peerAddr, procIdStr string, data interface{}) error {
	ctx := context.TODO()

	// Create a new libp2p host for this connection
	host, err := libp2p.New()
	if err != nil {
		return fmt.Errorf("failed to create libp2p host: %w", err)
	}
	defer host.Close()

	var peerInfo *peer.AddrInfo

	// Try to parse as peer ID first
	peerId, err := peer.Decode(peerAddr)
	if err == nil {
		// Successfully parsed as peer ID
		peerInfo = &peer.AddrInfo{
			ID: peerId,
			// Note: In a real implementation, you'd need peer discovery
			// or provide addresses as additional parameters
		}
	} else {
		// Fall back to treating as multiaddr
		addr, err := ma.NewMultiaddr(peerAddr)
		if err != nil {
			return fmt.Errorf("invalid peer address or ID: %w", err)
		}
		peerInfo, err = peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return fmt.Errorf("failed to parse peer info from multiaddr: %w", err)
		}
	}

	// Connect to the peer
	if err := host.Connect(ctx, *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	// Create protocol ID from process ID
	protocolID := protocol.ID("/ww/0.1.0/" + procIdStr)

	// Open stream to the peer
	stream, err := host.NewStream(ctx, peerInfo.ID, protocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
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
		return fmt.Errorf("unsupported data type: %T, expected io.Reader, []byte, or string", data)
	}

	// Send the data atomically
	_, err = io.Copy(stream, reader)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	return nil
}
