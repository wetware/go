package lang

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"capnproto.org/go/capnp/v3/exp/bufferpool"
	"github.com/ipfs/go-cid"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/system"
)

// IPFSObject wraps an IPFS capability and provides method access
type IPFSObject struct {
	IPFS system.IPFS
}

// Invoke implements core.Invokable to support direct calls like (IPFS "stat" path)
func (i *IPFSObject) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ipfs requires at least 1 argument: method name, got %d", len(args))
	}

	// First argument should be the method name
	var methodName string
	switch arg := args[0].(type) {
	case string:
		methodName = arg
	case builtin.String:
		methodName = string(arg)
	case builtin.Keyword:
		methodName = string(arg)
	default:
		return nil, fmt.Errorf("ipfs first argument must be string or keyword (method name), got %T", args[0])
	}

	// Call the appropriate method based on the name
	switch methodName {
	case "cat":
		return IPFSCat(append([]core.Any{i.IPFS}, args[1:]...)...)
	case "add":
		return IPFSAdd(append([]core.Any{i.IPFS}, args[1:]...)...)
	case "ls":
		return IPFSLs(append([]core.Any{i.IPFS}, args[1:]...)...)
	case "stat":
		return IPFSStat(append([]core.Any{i.IPFS}, args[1:]...)...)
	case "pin":
		return IPFSPin(append([]core.Any{i.IPFS}, args[1:]...)...)
	case "unpin":
		return IPFSUnpin(append([]core.Any{i.IPFS}, args[1:]...)...)
	case "pins":
		return IPFSPins(append([]core.Any{i.IPFS}, args[1:]...)...)
	case "id":
		return IPFSId(append([]core.Any{i.IPFS}, args[1:]...)...)
	case "connect":
		return IPFSConnect(append([]core.Any{i.IPFS}, args[1:]...)...)
	case "peers":
		return IPFSPeers(append([]core.Any{i.IPFS}, args[1:]...)...)
	default:
		return nil, fmt.Errorf("unknown IPFS method: %s", methodName)
	}
}

// Get implements core.Getter to support dot-method calls
func (i *IPFSObject) Get(key core.Any) (core.Any, error) {
	// Convert key to string
	var methodName string
	switch k := key.(type) {
	case builtin.String:
		methodName = string(k)
	case string:
		methodName = k
	case builtin.Keyword:
		methodName = string(k)
	default:
		return nil, fmt.Errorf("method name must be string or keyword, got %T", key)
	}

	// Return the appropriate method based on the name
	switch methodName {
	case "cat":
		return &IPFSInvokable{fn: func(args ...core.Any) (core.Any, error) {
			// Prepend the IPFS capability to the arguments
			return IPFSCat(append([]core.Any{i.IPFS}, args...)...)
		}}, nil
	case "add":
		return &IPFSInvokable{fn: func(args ...core.Any) (core.Any, error) {
			return IPFSAdd(append([]core.Any{i.IPFS}, args...)...)
		}}, nil
	case "ls":
		return &IPFSInvokable{fn: func(args ...core.Any) (core.Any, error) {
			return IPFSLs(append([]core.Any{i.IPFS}, args...)...)
		}}, nil
	case "stat":
		return &IPFSInvokable{fn: func(args ...core.Any) (core.Any, error) {
			return IPFSStat(append([]core.Any{i.IPFS}, args...)...)
		}}, nil
	case "pin":
		return &IPFSInvokable{fn: func(args ...core.Any) (core.Any, error) {
			return IPFSPin(append([]core.Any{i.IPFS}, args...)...)
		}}, nil
	case "unpin":
		return &IPFSInvokable{fn: func(args ...core.Any) (core.Any, error) {
			return IPFSUnpin(append([]core.Any{i.IPFS}, args...)...)
		}}, nil
	case "pins":
		return &IPFSInvokable{fn: func(args ...core.Any) (core.Any, error) {
			return IPFSPins(append([]core.Any{i.IPFS}, args...)...)
		}}, nil
	case "id":
		return &IPFSInvokable{fn: func(args ...core.Any) (core.Any, error) {
			return IPFSId(append([]core.Any{i.IPFS}, args...)...)
		}}, nil
	case "connect":
		return &IPFSInvokable{fn: func(args ...core.Any) (core.Any, error) {
			return IPFSConnect(append([]core.Any{i.IPFS}, args...)...)
		}}, nil
	case "peers":
		return &IPFSInvokable{fn: func(args ...core.Any) (core.Any, error) {
			return IPFSPeers(append([]core.Any{i.IPFS}, args...)...)
		}}, nil
	default:
		return nil, fmt.Errorf("unknown IPFS method: %s", methodName)
	}
}

// IPFSInvokable wraps an IPFS function to make it compatible with slurp's core.Invokable interface
type IPFSInvokable struct {
	fn func(args ...core.Any) (core.Any, error)
}

// Invoke implements core.Invokable for IPFSInvokable
func (i *IPFSInvokable) Invoke(args ...core.Any) (core.Any, error) {
	// The IPFS capability should be passed as the first argument from the RPC call
	return i.fn(args...)
}

// IPFSCat implements a standalone cat function for the shell
// Buffer wraps a *bytes.Buffer for use in the shell
type Buffer struct {
	Mem []byte
}

func (b *Buffer) Close() error {
	bufferpool.Default.Put(b.Mem)
	return nil
}

func (b Buffer) NewReader() *bytes.Reader {
	return bytes.NewReader(b.Mem)
}

func (b *Buffer) String() string {
	return string(b.Mem)
}

func (b *Buffer) SExpr() (string, error) {
	return b.AsHex(), nil
}

// AsHex returns a hex representation of the buffer contents
func (b *Buffer) AsHex() string {
	if len(b.Mem) == 0 {
		return "0x"
	}
	return "0x" + hex.EncodeToString(b.Mem)
}

// IPFSCat implements a standalone cat function for the shell
// Invoke implements core.Invokable for IPFSCat
func IPFSCat(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("cat requires at least 2 arguments: IPFS capability and path, got %d", len(args))
	}

	// First argument should be the IPFS capability
	ipfs, ok := args[0].(system.IPFS)
	if !ok {
		return nil, fmt.Errorf("cat first argument must be IPFS capability, got %T", args[0])
	}

	// Second argument should be the path
	var cid string
	switch arg := args[1].(type) {
	case string:
		// If it's a full path like "/ipfs/QmHash...", extract just the CID
		if strings.HasPrefix(arg, "/ipfs/") {
			cid = strings.TrimPrefix(arg, "/ipfs/")
		} else {
			cid = arg
		}
	case builtin.String:
		argStr := string(arg)
		if strings.HasPrefix(argStr, "/ipfs/") {
			cid = strings.TrimPrefix(argStr, "/ipfs/")
		} else {
			cid = argStr
		}
	case *UnixPath:
		// Extract CID from the path segments (e.g., ["ipfs", "QmHash..."] -> "QmHash...")
		segments := arg.Path.Segments()
		if len(segments) < 2 {
			return nil, fmt.Errorf("invalid IPFS path: insufficient segments")
		}
		cid = segments[1] // Get the CID part
	default:
		return nil, fmt.Errorf("cat second argument must be string or UnixPath, got %T", args[1])
	}

	// Call the Cat method
	ctx := context.Background()
	future, release := ipfs.Cat(ctx, func(params system.IPFS_cat_Params) error {
		return params.SetCid(cid)
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
	buf := bufferpool.Default.Get(len(body))
	copy(buf, body)

	return &Buffer{Mem: buf}, nil
}

// IPFSAdd adds data to IPFS
// Invoke implements core.Invokable for IPFSAdd
func IPFSAdd(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("add requires at least 2 arguments: IPFS capability and data, got %d", len(args))
	}

	// First argument should be the IPFS capability
	ipfs, ok := args[0].(system.IPFS)
	if !ok {
		return nil, fmt.Errorf("add first argument must be IPFS capability, got %T", args[0])
	}

	// Second argument should be the data
	buf, ok := args[1].(*Buffer)
	if !ok {
		return nil, fmt.Errorf("add second argument must be Buffer, got %T", args[1])
	}

	ctx := context.Background()
	future, release := ipfs.Add(ctx, func(params system.IPFS_add_Params) error {
		return params.SetData(buf.Mem)
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
// Invoke implements core.Invokable for IPFSLs
func IPFSLs(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("ls requires at least 2 arguments: IPFS capability and path, got %d", len(args))
	}

	// First argument should be the IPFS capability
	ipfs, ok := args[0].(system.IPFS)
	if !ok {
		return nil, fmt.Errorf("ls first argument must be IPFS capability, got %T", args[0])
	}

	// Second argument should be the path
	var pathStr string
	switch arg := args[1].(type) {
	case string:
		pathStr = arg
	case builtin.String:
		pathStr = string(arg)
	case *UnixPath:
		// Extract CID from the path segments (e.g., ["ipfs", "QmHash..."] -> "QmHash...")
		segments := arg.Path.Segments()
		if len(segments) < 2 {
			return nil, fmt.Errorf("invalid IPFS path: insufficient segments")
		}
		pathStr = segments[1] // Get the CID part
	default:
		return nil, fmt.Errorf("ls second argument must be string or UnixPath, got %T", args[1])
	}

	// Call the Ls method
	ctx := context.Background()
	future, release := ipfs.Ls(ctx, func(params system.IPFS_ls_Params) error {
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

	// Build a map of entries
	result := make(Map)
	for i := 0; i < entries.Len(); i++ {
		entry := entries.At(i)
		name, err := entry.Name()
		if err != nil {
			return nil, fmt.Errorf("failed to get entry name: %w", err)
		}

		cid, err := entry.Cid()
		if err != nil {
			return nil, fmt.Errorf("failed to get entry cid: %w", err)
		}

		// Create entry map with all available information
		// and store it under the name key.
		result[builtin.Keyword(name)] = Map{
			builtin.Keyword("name"): builtin.String(name),
			builtin.Keyword("cid"):  builtin.String(cid),
			builtin.Keyword("size"): builtin.Int64(entry.Size()),
			builtin.Keyword("type"): builtin.String(entry.Type().String()),
		}

	}
	return result, nil

}

// IPFSStat gets information about a CID
// Invoke implements core.Invokable for IPFSStat
func IPFSStat(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("stat requires at least 2 arguments: IPFS capability and cid, got %d", len(args))
	}

	// First argument should be the IPFS capability
	ipfs, ok := args[0].(system.IPFS)
	if !ok {
		return nil, fmt.Errorf("stat first argument must be IPFS capability, got %T", args[0])
	}

	// Second argument should be the CID
	var cid string
	switch arg := args[1].(type) {
	case string:
		// If it's a full path like "/ipfs/QmHash...", extract just the CID
		if strings.HasPrefix(arg, "/ipfs/") {
			cid = strings.TrimPrefix(arg, "/ipfs/")
		} else {
			cid = arg
		}
	case builtin.String:
		argStr := string(arg)
		if strings.HasPrefix(argStr, "/ipfs/") {
			cid = strings.TrimPrefix(argStr, "/ipfs/")
		} else {
			cid = argStr
		}
	case *UnixPath:
		// Extract CID from the path segments (e.g., ["ipfs", "QmHash..."] -> "QmHash...")
		segments := arg.Path.Segments()
		if len(segments) < 2 {
			return nil, fmt.Errorf("invalid IPFS path: insufficient segments")
		}
		cid = segments[1] // Get the CID part
	default:
		return nil, fmt.Errorf("stat second argument must be string or UnixPath, got %T", args[1])
	}

	ctx := context.Background()
	future, release := ipfs.Stat(ctx, func(params system.IPFS_stat_Params) error {
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
	infoCid, err := info.Cid()
	if err != nil {
		return nil, fmt.Errorf("failed to get cid: %w", err)
	}

	infoStr := fmt.Sprintf("CID: %s, Size: %d, Type: %s", infoCid, info.Size(), info.NodeType().Which())
	return builtin.String(infoStr), nil
}

// IPFSPin pins a CID
// Invoke implements core.Invokable for IPFSPin
func IPFSPin(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return builtin.Bool(false), fmt.Errorf("pin requires at least 2 arguments: IPFS capability and cid, got %d", len(args))
	}

	// First argument should be the IPFS capability
	ipfs, ok := args[0].(system.IPFS)
	if !ok {
		return builtin.Bool(false), fmt.Errorf("pin first argument must be IPFS capability, got %T", args[0])
	}

	// Second argument should be the CID
	var cid string
	switch arg := args[1].(type) {
	case string:
		cid = arg
	case builtin.String:
		cid = string(arg)
	case *UnixPath:
		cid = arg.String()
	default:
		return builtin.Bool(false), fmt.Errorf("pin second argument must be string or UnixPath, got %T", args[1])
	}

	ctx := context.Background()
	future, release := ipfs.Pin(ctx, func(params system.IPFS_pin_Params) error {
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
// Invoke implements core.Invokable for IPFSUnpin
func IPFSUnpin(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return builtin.Bool(false), fmt.Errorf("unpin requires at least 2 arguments: IPFS capability and cid, got %d", len(args))
	}

	// First argument should be the IPFS capability
	ipfs, ok := args[0].(system.IPFS)
	if !ok {
		return builtin.Bool(false), fmt.Errorf("unpin first argument must be IPFS capability, got %T", args[0])
	}

	// Second argument should be the CID
	var cid string
	switch arg := args[1].(type) {
	case string:
		cid = arg
	case builtin.String:
		cid = string(arg)
	case *UnixPath:
		cid = arg.String()
	default:
		return builtin.Bool(false), fmt.Errorf("unpin second argument must be string or UnixPath, got %T", args[1])
	}

	ctx := context.Background()
	future, release := ipfs.Unpin(ctx, func(params system.IPFS_unpin_Params) error {
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
// Invoke implements core.Invokable for IPFSPins
func IPFSPins(args ...core.Any) (core.Any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("pins requires at least 1 argument: IPFS capability, got %d", len(args))
	}

	// First argument should be the IPFS capability
	ipfs, ok := args[0].(system.IPFS)
	if !ok {
		return nil, fmt.Errorf("pins first argument must be IPFS capability, got %T", args[0])
	}

	ctx := context.Background()
	future, release := ipfs.Pins(ctx, func(params system.IPFS_pins_Params) error {
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
// Invoke implements core.Invokable for IPFSId
func IPFSId(args ...core.Any) (core.Any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("id requires at least 1 argument: IPFS capability, got %d", len(args))
	}

	// First argument should be the IPFS capability
	ipfs, ok := args[0].(system.IPFS)
	if !ok {
		return nil, fmt.Errorf("id first argument must be IPFS capability, got %T", args[0])
	}

	ctx := context.Background()
	future, release := ipfs.Id(ctx, func(params system.IPFS_id_Params) error {
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
// Invoke implements core.Invokable for IPFSConnect
func IPFSConnect(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("connect requires at least 2 arguments: IPFS capability and address, got %d", len(args))
	}

	// First argument should be the IPFS capability
	ipfs, ok := args[0].(system.IPFS)
	if !ok {
		return nil, fmt.Errorf("connect first argument must be IPFS capability, got %T", args[0])
	}

	// Second argument should be the address
	var addr string
	switch arg := args[1].(type) {
	case string:
		addr = arg
	case builtin.String:
		addr = string(arg)
	default:
		return nil, fmt.Errorf("connect second argument must be string, got %T", args[1])
	}

	ctx := context.Background()
	future, release := ipfs.Connect(ctx, func(params system.IPFS_connect_Params) error {
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
// Invoke implements core.Invokable for IPFSPeers
func IPFSPeers(args ...core.Any) (core.Any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("peers requires at least 1 argument: IPFS capability, got %d", len(args))
	}

	// First argument should be the IPFS capability
	ipfs, ok := args[0].(system.IPFS)
	if !ok {
		return nil, fmt.Errorf("peers first argument must be IPFS capability, got %T", args[0])
	}

	ctx := context.Background()
	future, release := ipfs.Peers(ctx, func(params system.IPFS_peers_Params) error {
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

// Go implements the go special form for spawning processes across different contexts
type Go struct {
	Executor system.Executor
}

// Invoke implements core.Invokable for Go
func (g Go) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("go requires at least 2 arguments: executable path and body, got %d", len(args))
	}

	// First argument should be the executable path
	execPath, ok := args[0].(*UnixPath)
	if !ok {
		return nil, fmt.Errorf("go first argument must be UnixPath (executable path), got %T", args[0])
	}

	// Second argument should be the body (code to execute)
	body := args[1] // Already core.Any, but may have methods in the future

	// Parse keyword arguments (optional)
	kwargs := make(map[string]core.Any)
	for i := 2; i < len(args); i++ {
		if i+1 >= len(args) {
			return nil, fmt.Errorf("go keyword argument at position %d has no value", i)
		}

		key, ok := args[i].(builtin.String)
		if !ok {
			return nil, fmt.Errorf("go keyword argument key must be string, got %T", args[i])
		}

		kwargs[string(key)] = args[i+1]
		i++ // Skip the value in next iteration
	}

	// Extract executor from kwargs or use default
	if execKwarg, exists := kwargs["exec"]; exists {
		execCap, ok := execKwarg.(system.Executor)
		if !ok {
			return nil, fmt.Errorf("go :exec keyword argument must be Executor capability, got %T", execKwarg)
		}
		_ = execCap // TODO: Use executor in future implementation
	}

	// Extract console from kwargs if provided
	if consoleKwarg, exists := kwargs["console"]; exists {
		_ = consoleKwarg // TODO: Use console in future implementation
	}

	// Extract other capabilities from kwargs
	for key, value := range kwargs {
		if key == "exec" || key == "console" {
			continue // Already handled
		}

		// Convert capability to CapDescriptor
		// TODO: Implement proper capability conversion
		_ = key
		_ = value
	}

	// Serialize the body to a string representation
	bodyStr, err := serializeBody(body)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize body: %w", err)
	}

	// Prepare arguments for the executor
	var cmdArgs []string
	cmdArgs = append(cmdArgs, bodyStr)

	// Add any additional arguments from kwargs
	for key, value := range kwargs {
		if key == "exec" || key == "console" {
			continue // Already handled
		}
		// Convert value to string and add as argument
		argStr := fmt.Sprintf("--%s=%v", key, value)
		cmdArgs = append(cmdArgs, argStr)
	}

	// Spawn the process using the executor
	ctx := context.Background()
	future, release := g.Executor.Spawn(ctx, func(call system.Executor_spawn_Params) error {
		params, err := call.NewCommand()
		if err != nil {
			return fmt.Errorf("failed to create command: %w", err)
		}

		// Set the executable path
		if err := params.SetPath(execPath.String()); err != nil {
			return fmt.Errorf("failed to set path: %w", err)
		}

		// Set arguments
		argsList, err := params.NewArgs(int32(len(cmdArgs)))
		if err != nil {
			return fmt.Errorf("failed to create args list: %w", err)
		}
		for i, arg := range cmdArgs {
			if err := argsList.Set(i, arg); err != nil {
				return fmt.Errorf("failed to set arg %d: %w", i, err)
			}
		}
		if err := params.SetArgs(argsList); err != nil {
			return fmt.Errorf("failed to set args: %w", err)
		}

		// Set environment variables (empty for now, could be enhanced)
		envList, err := params.NewEnv(0)
		if err != nil {
			return fmt.Errorf("failed to create env list: %w", err)
		}
		if err := params.SetEnv(envList); err != nil {
			return fmt.Errorf("failed to set env: %w", err)
		}

		// Set working directory (empty for now, could be enhanced)
		if err := params.SetDir(""); err != nil {
			return fmt.Errorf("failed to set dir: %w", err)
		}

		return nil
	})
	defer release()

	// Get the result
	result, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get spawn result: %w", err)
	}

	// Get the cell
	optionalCell, err := result.Cell()
	if err != nil {
		return nil, fmt.Errorf("failed to get cell: %w", err)
	}

	// Check if spawn was successful
	if optionalCell.Which() == system.Executor_OptionalCell_Which_err {
		errStruct := optionalCell.Err()
		status := errStruct.Status()
		body, err := errStruct.Body()
		if err != nil {
			return nil, fmt.Errorf("failed to get body: %w", err)
		}
		return nil, fmt.Errorf("spawn failed with status %d: %s", status, string(body))
	}

	// Get the cell
	cell := optionalCell.Cell()
	if !cell.IsValid() {
		return nil, fmt.Errorf("spawn returned invalid cell")
	}

	// Return the cell for further interaction
	return cell, nil
}

// serializeBody converts the body to a string representation
func serializeBody(body core.Any) (string, error) {
	// Convert body to a string representation
	var bodyStr string

	switch v := body.(type) {
	case builtin.String:
		bodyStr = string(v)
	case interface {
		Len() int
		At(int) core.Any
	}:
		// Convert list-like object to s-expression string
		items := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			item := v.At(i)
			if str, ok := item.(builtin.String); ok {
				items[i] = fmt.Sprintf("%q", string(str))
			} else {
				items[i] = fmt.Sprintf("%v", item)
			}
		}
		bodyStr = fmt.Sprintf("(%s)", strings.Join(items, " "))
	default:
		bodyStr = fmt.Sprintf("%v", v)
	}

	return bodyStr, nil
}

// DotNotationAnalyzer implements a custom analyzer that handles dot notation
type DotNotationAnalyzer struct {
	base core.Analyzer
}

// NewDotNotationAnalyzer creates a new analyzer that handles dot notation
func NewDotNotationAnalyzer(base core.Analyzer) *DotNotationAnalyzer {
	if base == nil {
		base = &builtin.Analyzer{}
	}

	return &DotNotationAnalyzer{base: base}
}

// Analyze implements core.Analyzer and handles dot notation expressions.
//
// This function performs syntactic analysis of forms and converts dot notation
// expressions into standard function calls. It works by:
//
// 1. Detecting when a list's first element is a symbol containing a dot (e.g., "ipfs.stat")
// 2. Splitting the symbol into object and method parts (e.g., "ipfs" and "stat")
// 3. Converting the expression from (ipfs.stat arg1 arg2) to (ipfs "stat" arg1 arg2)
// 4. Delegating the transformed expression to the base analyzer
//
// Examples of transformations:
//
//	(ipfs.stat path)     → (ipfs "stat" path)
//	(ipfs.cat cid)       → (ipfs "cat" cid)
//	(ipfs.id)            → (ipfs "id")
//	(obj.method a b c)   → (obj "method" a b c)
//
// The function preserves all arguments after the dot notation symbol and
// maintains the original evaluation order. If the form is not a dot notation
// expression, it delegates directly to the base analyzer without modification.
func (dna *DotNotationAnalyzer) Analyze(env core.Env, form core.Any) (core.Expr, error) {
	// Step 1: Check if the form is a sequence (list) that might contain dot notation
	// We only process sequences since dot notation only makes sense in function calls
	if seq, ok := form.(core.Seq); ok {
		// Get the count of elements in the sequence
		cnt, err := seq.Count()
		if err != nil {
			return nil, fmt.Errorf("failed to count sequence elements: %w", err)
		}

		// Only process non-empty sequences (empty lists can't have dot notation)
		if cnt > 0 {
			// Step 2: Extract the first element to check if it's a dot notation symbol
			first, err := seq.First()
			if err != nil {
				return nil, fmt.Errorf("failed to get first sequence element: %w", err)
			}

			// Step 3: Check if the first element is a symbol that contains a dot
			// Dot notation only applies to symbols like "ipfs.stat", "obj.method", etc.
			if sym, ok := first.(builtin.Symbol); ok {
				symStr := string(sym)

				// Look for the dot character that indicates method notation
				if strings.Contains(symStr, ".") {
					// Step 4: Split the symbol into object and method parts
					// We use SplitN with limit 2 to handle only the first dot
					// This allows for future extensions like "ipfs.stat.detail" if needed
					parts := strings.SplitN(symStr, ".", 2)
					if len(parts) == 2 {
						// Extract the object name (e.g., "ipfs") and method name (e.g., "stat")
						object := builtin.Symbol(parts[0]) // The object to call the method on
						method := builtin.String(parts[1]) // The method name as a string

						// Step 5: Collect all arguments for the transformed function call
						// Start with the object and method name
						var args []core.Any
						args = append(args, object, method)

						// Step 6: Iterate through the remaining arguments in the original sequence
						// We need to preserve all arguments that came after the dot notation symbol
						current := seq
						for i := 1; i < int(cnt); i++ {
							// Get the next element in the sequence
							next, err := current.Next()
							if err != nil {
								return nil, fmt.Errorf("failed to get next sequence element at position %d: %w", i, err)
							}
							if next == nil {
								// End of sequence reached
								break
							}

							// Extract the actual argument value from the sequence
							arg, err := next.First()
							if err != nil {
								return nil, fmt.Errorf("failed to get argument at position %d: %w", i, err)
							}

							// Add the argument to our transformed function call
							args = append(args, arg)

							// Move to the next element in the sequence
							current = next
						}

						// Step 7: Create a new list representing the transformed function call
						// This converts (ipfs.stat path) into (ipfs "stat" path)
						newForm := builtin.NewList(args...)

						// Step 8: Delegate the transformed expression to the base analyzer
						// The base analyzer will handle the actual evaluation of the function call
						return dna.base.Analyze(env, newForm)
					}
					// If splitting didn't produce exactly 2 parts, it's not valid dot notation
					// Fall through to base analyzer which will likely produce an error
				}
			}
		}
	}

	// Step 9: If the form is not a dot notation expression, delegate to base analyzer
	// This handles all other forms (regular function calls, literals, etc.) normally
	return dna.base.Analyze(env, form)
}
