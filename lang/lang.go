package lang

import (
	"context"
	"fmt"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/system"
)

var _ core.Invokable = (*Session)(nil)
var _ core.Env = (*Session)(nil)

type Session struct {
	IPFS system.IPFS
}

// Cat retrieves data from IPFS by CID
func (s Session) Cat(cid string) ([]byte, error) {
	ctx := context.Background()
	future, release := s.IPFS.Cat(ctx, func(params system.IPFS_cat_Params) error {
		return params.SetCid(cid)
	})
	defer release()

	results, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get cat results: %w", err)
	}

	body, err := results.Body()
	if err != nil {
		return nil, fmt.Errorf("failed to get body: %w", err)
	}

	// Copy the data to avoid memory management issues
	result := make([]byte, len(body))
	copy(result, body)

	return result, nil
}

// Add adds data to IPFS
func (s Session) Add(data []byte) (string, error) {
	ctx := context.Background()
	future, release := s.IPFS.Add(ctx, func(params system.IPFS_add_Params) error {
		return params.SetData(data)
	})
	defer release()

	results, err := future.Struct()
	if err != nil {
		return "", fmt.Errorf("failed to get add results: %w", err)
	}

	cid, err := results.Cid()
	if err != nil {
		return "", fmt.Errorf("failed to get CID: %w", err)
	}

	return cid, nil
}

// Ls lists contents of a directory or object
func (s Session) Ls(path string) ([]system.Entry, error) {
	ctx := context.Background()
	future, release := s.IPFS.Ls(ctx, func(params system.IPFS_ls_Params) error {
		return params.SetPath(path)
	})
	defer release()

	results, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get ls results: %w", err)
	}

	entries, err := results.Entries()
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// Convert to slice
	var result []system.Entry
	for i := 0; i < entries.Len(); i++ {
		result = append(result, entries.At(i))
	}

	return result, nil
}

// Stat gets information about a CID
func (s Session) Stat(cid string) (*system.NodeInfo, error) {
	ctx := context.Background()
	future, release := s.IPFS.Stat(ctx, func(params system.IPFS_stat_Params) error {
		return params.SetCid(cid)
	})
	defer release()

	results, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get stat results: %w", err)
	}

	info, err := results.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	return &info, nil
}

// Pin pins a CID
func (s Session) Pin(cid string) (bool, error) {
	ctx := context.Background()
	future, release := s.IPFS.Pin(ctx, func(params system.IPFS_pin_Params) error {
		return params.SetCid(cid)
	})
	defer release()

	results, err := future.Struct()
	if err != nil {
		return false, fmt.Errorf("failed to get pin results: %w", err)
	}

	return results.Success(), nil
}

// Unpin unpins a CID
func (s Session) Unpin(cid string) (bool, error) {
	ctx := context.Background()
	future, release := s.IPFS.Unpin(ctx, func(params system.IPFS_unpin_Params) error {
		return params.SetCid(cid)
	})
	defer release()

	results, err := future.Struct()
	if err != nil {
		return false, fmt.Errorf("failed to get unpin results: %w", err)
	}

	return results.Success(), nil
}

// Pins lists pinned CIDs
func (s Session) Pins() ([]string, error) {
	ctx := context.Background()
	future, release := s.IPFS.Pins(ctx, func(params system.IPFS_pins_Params) error {
		return nil // no parameters needed
	})
	defer release()

	results, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get pins results: %w", err)
	}

	cids, err := results.Cids()
	if err != nil {
		return nil, fmt.Errorf("failed to get CIDs: %w", err)
	}

	// Convert to slice
	var result []string
	for i := 0; i < cids.Len(); i++ {
		cid, err := cids.At(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get CID at index %d: %w", i, err)
		}
		result = append(result, cid)
	}

	return result, nil
}

// Id gets peer information
func (s Session) Id() (*system.PeerInfo, error) {
	ctx := context.Background()
	future, release := s.IPFS.Id(ctx, func(params system.IPFS_id_Params) error {
		return nil // no parameters needed
	})
	defer release()

	results, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get id results: %w", err)
	}

	info, err := results.PeerInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get peer info: %w", err)
	}

	return &info, nil
}

// Connect connects to a peer
func (s Session) Connect(addr string) (bool, error) {
	ctx := context.Background()
	future, release := s.IPFS.Connect(ctx, func(params system.IPFS_connect_Params) error {
		return params.SetAddr(addr)
	})
	defer release()

	results, err := future.Struct()
	if err != nil {
		return false, fmt.Errorf("failed to get connect results: %w", err)
	}

	return results.Success(), nil
}

// Peers lists connected peers
func (s Session) Peers() ([]system.PeerInfo, error) {
	ctx := context.Background()
	future, release := s.IPFS.Peers(ctx, func(params system.IPFS_peers_Params) error {
		return nil // no parameters needed
	})
	defer release()

	results, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get peers results: %w", err)
	}

	peerList, err := results.PeerList()
	if err != nil {
		return nil, fmt.Errorf("failed to get peer list: %w", err)
	}

	// Convert to slice
	var result []system.PeerInfo
	for i := 0; i < peerList.Len(); i++ {
		result = append(result, peerList.At(i))
	}

	return result, nil
}

// Invoke implements core.Invokable interface for Session
func (s Session) Invoke(args ...core.Any) (core.Any, error) {
	// If no arguments are provided, return the wrapper itself
	if len(args) == 0 {
		return s, nil
	}

	// The first argument must be a string containing the method name
	var methodName string
	switch arg := args[0].(type) {
	case string:
		methodName = arg
	case builtin.String:
		methodName = string(arg)
	default:
		return nil, fmt.Errorf("first argument must be method name string, got %T", args[0])
	}

	// Call the appropriate method based on the method name
	switch methodName {
	case "Cat":
		if len(args) != 2 {
			return nil, fmt.Errorf("Cat requires exactly 1 argument, got %d", len(args)-1)
		}
		var cid string
		switch arg := args[1].(type) {
		case string:
			cid = arg
		case builtin.String:
			cid = string(arg)
		default:
			return nil, fmt.Errorf("Cat argument must be string, got %T", args[1])
		}
		return s.Cat(cid)

	case "Add":
		if len(args) != 2 {
			return nil, fmt.Errorf("Add requires exactly 1 argument, got %d", len(args)-1)
		}
		var data []byte
		switch arg := args[1].(type) {
		case []byte:
			data = arg
		case builtin.String:
			data = []byte(arg)
		case string:
			data = []byte(arg)
		default:
			return nil, fmt.Errorf("Add argument must be string or []byte, got %T", args[1])
		}
		return s.Add(data)

	case "Ls":
		if len(args) != 2 {
			return nil, fmt.Errorf("Ls requires exactly 1 argument, got %d", len(args)-1)
		}
		var path string
		switch arg := args[1].(type) {
		case string:
			path = arg
		case builtin.String:
			path = string(arg)
		default:
			return nil, fmt.Errorf("Ls argument must be string, got %T", args[1])
		}
		return s.Ls(path)

	case "Stat":
		if len(args) != 2 {
			return nil, fmt.Errorf("Stat requires exactly 1 argument, got %d", len(args)-1)
		}
		var cid string
		switch arg := args[1].(type) {
		case string:
			cid = arg
		case builtin.String:
			cid = string(arg)
		default:
			return nil, fmt.Errorf("Stat argument must be string, got %T", args[1])
		}
		return s.Stat(cid)

	case "Pin":
		if len(args) != 2 {
			return nil, fmt.Errorf("Pin requires exactly 1 argument, got %d", len(args)-1)
		}
		var cid string
		switch arg := args[1].(type) {
		case string:
			cid = arg
		case builtin.String:
			cid = string(arg)
		default:
			return nil, fmt.Errorf("Pin argument must be string, got %T", args[1])
		}
		return s.Pin(cid)

	case "Unpin":
		if len(args) != 2 {
			return nil, fmt.Errorf("Unpin requires exactly 1 argument, got %d", len(args)-1)
		}
		var cid string
		switch arg := args[1].(type) {
		case string:
			cid = arg
		case builtin.String:
			cid = string(arg)
		default:
			return nil, fmt.Errorf("Unpin argument must be string, got %T", args[1])
		}
		return s.Unpin(cid)

	case "Pins":
		if len(args) != 1 {
			return nil, fmt.Errorf("Pins requires no arguments, got %d", len(args)-1)
		}
		return s.Pins()

	case "Id":
		if len(args) != 1 {
			return nil, fmt.Errorf("Id requires no arguments, got %d", len(args)-1)
		}
		return s.Id()

	case "Connect":
		if len(args) != 2 {
			return nil, fmt.Errorf("Connect requires exactly 1 argument, got %d", len(args)-1)
		}
		var addr string
		switch arg := args[1].(type) {
		case string:
			addr = arg
		case builtin.String:
			addr = string(arg)
		default:
			return nil, fmt.Errorf("Connect argument must be string, got %T", args[1])
		}
		return s.Connect(addr)

	case "Peers":
		if len(args) != 1 {
			return nil, fmt.Errorf("Peers requires no arguments, got %d", len(args)-1)
		}
		return s.Peers()

	default:
		return nil, fmt.Errorf("unknown method: %s", methodName)
	}
}

// MethodWrapper wraps a Session method to make it invokable
type MethodWrapper struct {
	session Session
	method  string
}

// Invoke implements core.Invokable for MethodWrapper
func (mw MethodWrapper) Invoke(args ...core.Any) (core.Any, error) {
	// Prepend the method name as the first argument
	allArgs := append([]core.Any{mw.method}, args...)
	return mw.session.Invoke(allArgs...)
}

func (s Session) Bind(name string, val core.Any) error {
	return nil
}

func (s Session) Resolve(name string) (core.Any, error) {
	// Return a method wrapper for the requested method
	return MethodWrapper{session: s, method: name}, nil
}

func (s Session) Child(name string, vars map[string]core.Any) core.Env {
	return s
}

func (s Session) Name() string {
	return "ipfs"
}

func (s Session) Parent() core.Env {
	return nil
}
