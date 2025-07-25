package lang

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/system"
)

// IPFSCat implements a standalone cat function for the shell
// Buffer wraps a *bytes.Buffer for use in the shell
type Buffer struct {
	bytes.Buffer
}

// String returns a hex representation of the buffer contents
func (b *Buffer) String() string {
	if b.Buffer.Len() == 0 {
		return "0x"
	}
	return "0x" + hex.EncodeToString(b.Buffer.Bytes())
}

type IPFSCat struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable for IPFSCat
func (ic IPFSCat) Invoke(args ...core.Any) (core.Any, error) {
	// Identity law: when called with no arguments, return self
	if len(args) == 0 {
		return ic, nil
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("cat requires exactly 1 argument, got %d", len(args))
	}

	// Extract the path from the argument
	unixPath, ok := args[0].(*UnixPath)
	if !ok {
		return nil, fmt.Errorf("cat argument must be UnixPath, got %T", args[0])
	}

	// Call the Cat method
	ctx := context.Background()
	future, release := ic.IPFS.Cat(ctx, func(params system.IPFS_cat_Params) error {
		// Extract CID from the path segments (e.g., ["ipfs", "QmHash..."] -> "QmHash...")
		segments := unixPath.Path.Segments()
		if len(segments) < 2 {
			return fmt.Errorf("invalid IPFS path: insufficient segments")
		}
		return params.SetCid(segments[1])
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get cat results: %w", err)
	}

	body, err := res.Body()
	if err != nil {
		return nil, fmt.Errorf("failed to get body: %w", err)
	}

	// Create a buffer with the data
	buffer := &Buffer{}
	buffer.Write(body)

	return buffer, nil
}

// IPFSAdd adds data to IPFS
type IPFSAdd struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable for IPFSAdd
func (ia IPFSAdd) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("add requires exactly 1 argument, got %d", len(args))
	}

	// Extract the data from the argument
	buf, ok := args[0].(*Buffer)
	if !ok {
		return nil, fmt.Errorf("add argument must be Buffer, got %T", args[0])
	}

	ctx := context.Background()
	future, release := ia.IPFS.Add(ctx, func(params system.IPFS_add_Params) error {
		return params.SetData(buf.Bytes())
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return "", fmt.Errorf("failed to get add results: %w", err)
	}

	cid, err := res.Cid()
	if err != nil {
		return "", fmt.Errorf("failed to get CID: %w", err)
	}

	return cid, nil
}

// IPFSLs lists contents of a directory or object
type IPFSLs struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable for IPFSLs
func (il *IPFSLs) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ls requires exactly 1 argument, got %d", len(args))
	}

	// Extract the path from the argument
	var pathStr string
	switch arg := args[0].(type) {
	case string:
		pathStr = arg
	case builtin.String:
		pathStr = string(arg)
	case *UnixPath:
		pathStr = arg.String()
	default:
		return nil, fmt.Errorf("ls argument must be string or UnixPath, got %T", args[0])
	}

	// Call the Ls method
	ctx := context.Background()
	future, release := il.IPFS.Ls(ctx, func(params system.IPFS_ls_Params) error {
		return params.SetPath(pathStr)
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get ls results: %w", err)
	}

	entries, err := res.Entries()
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	var result []core.Any
	for i := 0; i < entries.Len(); i++ {
		entry := entries.At(i)
		result = append(result, builtin.String(entry.String()))
	}

	return builtin.NewList(result...), nil
}

// IPFSStat gets information about a CID
type IPFSStat struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable for IPFSStat
func (is *IPFSStat) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("stat requires exactly 1 argument, got %d", len(args))
	}

	// Extract the CID from the argument
	var cid string
	switch arg := args[0].(type) {
	case string:
		cid = arg
	case builtin.String:
		cid = string(arg)
	case *UnixPath:
		cid = arg.String()
	default:
		return nil, fmt.Errorf("stat argument must be string or UnixPath, got %T", args[0])
	}

	ctx := context.Background()
	future, release := is.IPFS.Stat(ctx, func(params system.IPFS_stat_Params) error {
		return params.SetCid(cid)
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get stat results: %w", err)
	}

	info, err := res.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	// Convert info to a string representation for shell usability
	infoCid, _ := info.Cid()
	infoType, _ := info.Type()
	infoStr := fmt.Sprintf("CID: %s, Size: %d, Type: %s", infoCid, info.Size(), infoType)
	return builtin.String(infoStr), nil
}

// IPFSPin pins a CID
type IPFSPin struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable for IPFSPin
func (ip *IPFSPin) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return builtin.Bool(false), fmt.Errorf("pin requires exactly 1 argument, got %d", len(args))
	}

	// Extract the CID from the argument
	var cid string
	switch arg := args[0].(type) {
	case string:
		cid = arg
	case builtin.String:
		cid = string(arg)
	case *UnixPath:
		cid = arg.String()
	default:
		return builtin.Bool(false), fmt.Errorf("pin argument must be string or UnixPath, got %T", args[0])
	}

	ctx := context.Background()
	future, release := ip.IPFS.Pin(ctx, func(params system.IPFS_pin_Params) error {
		return params.SetCid(cid)
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return builtin.Bool(false), fmt.Errorf("failed to get pin results: %w", err)
	}

	return builtin.Bool(res.Success()), nil
}

// IPFSUnpin unpins a CID
type IPFSUnpin struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable for IPFSUnpin
func (iu *IPFSUnpin) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return builtin.Bool(false), fmt.Errorf("unpin requires exactly 1 argument, got %d", len(args))
	}

	// Extract the CID from the argument
	var cid string
	switch arg := args[0].(type) {
	case string:
		cid = arg
	case builtin.String:
		cid = string(arg)
	case *UnixPath:
		cid = arg.String()
	default:
		return builtin.Bool(false), fmt.Errorf("unpin argument must be string or UnixPath, got %T", args[0])
	}

	ctx := context.Background()
	future, release := iu.IPFS.Unpin(ctx, func(params system.IPFS_unpin_Params) error {
		return params.SetCid(cid)
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return builtin.Bool(false), fmt.Errorf("failed to get unpin results: %w", err)
	}

	return builtin.Bool(res.Success()), nil
}

// IPFSPins lists pinned CIDs
type IPFSPins struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable for IPFSPins
func (ips *IPFSPins) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("pins requires no arguments, got %d", len(args))
	}

	ctx := context.Background()
	future, release := ips.IPFS.Pins(ctx, func(params system.IPFS_pins_Params) error {
		return nil
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get pins results: %w", err)
	}

	cids, err := res.Cids()
	if err != nil {
		return nil, fmt.Errorf("failed to get cids: %w", err)
	}

	// Convert CIDs to a list for shell usability
	var result []core.Any
	for i := 0; i < cids.Len(); i++ {
		rawCID, err := cids.At(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get cid: %w", err)
		}

		c, err := cid.Decode(rawCID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cid: %w", err)
		}

		result = append(result, c)
	}

	return builtin.NewList(result...), nil
}

// IPFSId gets peer information
type IPFSId struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable for IPFSId
func (ii *IPFSId) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("id requires no arguments, got %d", len(args))
	}

	ctx := context.Background()
	future, release := ii.IPFS.Id(ctx, func(params system.IPFS_id_Params) error {
		return nil
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get id results: %w", err)
	}

	info, err := res.PeerInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	// Convert info to a string representation for shell usability
	peerId, err := info.Id()
	if err != nil {
		return nil, fmt.Errorf("failed to get peer id: %w", err)
	}

	idStr := fmt.Sprintf("Peer ID: %s", peerId)
	return builtin.String(idStr), nil
}

// IPFSConnect connects to a peer
type IPFSConnect struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable for IPFSConnect
func (ic *IPFSConnect) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("connect requires exactly 1 argument, got %d", len(args))
	}

	// Extract the address from the argument
	var addr string
	switch arg := args[0].(type) {
	case string:
		addr = arg
	case builtin.String:
		addr = string(arg)
	default:
		return nil, fmt.Errorf("connect argument must be string, got %T", args[0])
	}

	ctx := context.Background()
	future, release := ic.IPFS.Connect(ctx, func(params system.IPFS_connect_Params) error {
		return params.SetAddr(addr)
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get connect results: %w", err)
	}

	return builtin.Bool(res.Success()), nil
}

// IPFSPeers lists connected peers
type IPFSPeers struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable for IPFSPeers
func (ip *IPFSPeers) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("peers requires no arguments, got %d", len(args))
	}

	ctx := context.Background()
	future, release := ip.IPFS.Peers(ctx, func(params system.IPFS_peers_Params) error {
		return nil
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get peers results: %w", err)
	}

	peers, err := res.PeerList()
	if err != nil {
		return nil, fmt.Errorf("failed to get peers: %w", err)
	}

	var result []core.Any
	for i := 0; i < peers.Len(); i++ {
		peer := peers.At(i)
		result = append(result, builtin.String(peer.String()))
	}

	return builtin.NewList(result...), nil
}
