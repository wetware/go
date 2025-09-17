package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/system"
)

var _ core.Invokable = (*Exec)(nil)

type Exec struct {
	Session interface {
		Exec() system.Executor
	}
}

//	  (exec <path>
//		  :timeout 15s)
func (e Exec) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("exec requires at least one argument (bytecode or reader)")
	}

	p, ok := args[0].(path.Path)
	if !ok {
		return nil, fmt.Errorf("exec expects a path, got %T", args[0])
	}

	// Process remaining args as key-value pairs
	opts := make(map[builtin.Keyword]core.Any)
	for i := 1; i < len(args); i += 2 {
		key, ok := args[i].(builtin.Keyword)
		if !ok {
			return nil, fmt.Errorf("option key must be a keyword, got %T", args[i])
		}

		if i+1 >= len(args) {
			return nil, fmt.Errorf("missing value for option %s", key)
		}

		opts[key] = args[i+1]
	}
	ctx, cancel := e.NewContext(opts)
	defer cancel()

	n, err := env.IPFS.Unixfs().Get(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve node %v: %w", p, err)
	}

	switch node := n.(type) {
	case files.File:
		bytecode, err := io.ReadAll(node)
		if err != nil {
			return nil, fmt.Errorf("failed to read bytecode: %w", err)
		}

		procID, err := e.ExecBytes(ctx, bytecode)
		if err != nil {
			return nil, fmt.Errorf("failed to execute bytecode: %w", err)
		}
		return builtin.String(procID), nil

	case files.Directory:
		return nil, errors.New("TODO:  directory support")
	default:
		return nil, fmt.Errorf("unexpected node type: %T", node)
	}
}

func (e Exec) ExecBytes(ctx context.Context, bytecode []byte) (string, error) {
	f, release := e.Session.Exec().Exec(ctx, func(p system.Executor_exec_Params) error {
		return p.SetBytecode(bytecode)
	})
	defer release()

	// Wait for the protocol setup to complete
	result, err := f.Struct()
	if err != nil {
		return "", err
	}

	procID, err := result.Protocol()
	return procID, err
}

func (e Exec) NewContext(opts map[builtin.Keyword]core.Any) (context.Context, context.CancelFunc) {
	// TODO:  add support for parsing durations like 15s, 15m, 15h, 15d
	// if timeout, ok := opts["timeout"].(time.Duration); ok {
	// 	return context.WithTimeout(context.Background(), timeout)
	// }

	return context.WithTimeout(context.Background(), time.Second*15)
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
