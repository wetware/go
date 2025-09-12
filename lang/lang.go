package lang

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/system"
)

func NewExecutor(ctx context.Context, client system.Executor) core.Any {
	return &Executor{Client: client}
}

func NewIPFS(ctx context.Context, ipfs iface.CoreAPI) core.Any {
	return &IPFS{CoreAPI: ipfs}
}

var _ core.Invokable = (*Executor)(nil)
var _ core.Invokable = (*IPFS)(nil)
var _ core.Invokable = (*Path)(nil)

type Executor struct {
	Client system.Executor
}

//	  (exec buffer
//		  :timeout 15s)
func (e Executor) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("exec requires at least one argument (bytecode or reader)")
	}

	var bytecode []byte
	var err error
	switch src := args[0].(type) {
	case []byte:
		bytecode = src
	case io.Reader:
		bytecode, err = io.ReadAll(src)
	default:
		err = fmt.Errorf("exec expects a reader or string, got %T", args[0])
	}
	if err != nil {
		return nil, err
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

	future, release := e.Client.Exec(ctx, func(p system.Executor_exec_Params) error {
		return p.SetBytecode(bytecode)
	})
	defer release()

	result, err := future.Struct()
	if err != nil {
		return "", err
	}

	protocol, err := result.Protocol()
	return builtin.String(protocol), err
}

func (e Executor) NewContext(opts map[builtin.Keyword]core.Any) (context.Context, context.CancelFunc) {
	if timeout, ok := opts["timeout"].(time.Duration); ok {
		return context.WithTimeout(context.Background(), timeout)
	}

	return context.WithTimeout(context.Background(), time.Second*15)
}

type IPFS struct {
	iface.CoreAPI
}

// IPFS methods: (ipfs :cat /ipfs/Qm...) or (ipfs :get /ipfs/Qm...)
func (i IPFS) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("ipfs requires at least 2 arguments: (ipfs :method path)")
	}

	// First argument should be a keyword (:cat or :get)
	method, ok := args[0].(builtin.Keyword)
	if !ok {
		return nil, fmt.Errorf("ipfs method must be a keyword, got %T", args[0])
	}

	// Second argument should be the IPFS path (string or Path object)
	var ipfsPath path.Path
	var err error

	switch p := args[1].(type) {
	case builtin.String:
		// Parse the IPFS path from string
		ipfsPath, err = path.NewPath(string(p))
		if err != nil {
			return nil, fmt.Errorf("invalid IPFS path %s: %w", p, err)
		}
	case *Path:
		// Use the path from Path object
		ipfsPath = p.Path
	default:
		return nil, fmt.Errorf("ipfs path must be a string or Path object, got %T", args[1])
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	switch method {
	case "cat":
		return i.Cat(ctx, ipfsPath)
	case "get":
		return i.Get(ctx, ipfsPath)
	default:
		return nil, fmt.Errorf("unknown ipfs method: %s (supported: :cat, :get)", method)
	}
}

// Cat returns the content of an IPFS file as []byte
func (i IPFS) Cat(ctx context.Context, p path.Path) (core.Any, error) {
	if i.CoreAPI == nil || i.CoreAPI.Unixfs() == nil {
		return nil, fmt.Errorf("IPFS client not initialized")
	}

	// Get the node from IPFS
	node, err := i.Unixfs().Get(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPFS path: %w", err)
	}

	// Handle different node types
	switch node := node.(type) {
	case files.File:
		// For files, read the content into a byte slice
		content, err := io.ReadAll(node)
		if err != nil {
			return nil, fmt.Errorf("failed to read file content: %w", err)
		}
		return content, nil
	case files.Directory:
		return nil, fmt.Errorf("path is a directory, use :get instead of :cat")
	default:
		return nil, fmt.Errorf("unexpected node type: %T", node)
	}
}

// Get returns the IPFS node as a file-like object (io.Reader) that can be used with eval
func (i IPFS) Get(ctx context.Context, p path.Path) (core.Any, error) {
	if i.CoreAPI == nil || i.CoreAPI.Unixfs() == nil {
		return nil, fmt.Errorf("IPFS client not initialized")
	}

	// Get the node from IPFS
	node, err := i.Unixfs().Get(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPFS path: %w", err)
	}

	switch node := node.(type) {
	case files.File:
		return File{File: node}, nil
	case files.Directory:
		return Directory{Directory: node}, nil
	default:
		return Node{Node: node}, nil
	}
}

// Path methods: (/ipfs/Qm... :cat) or (/ipfs/Qm...)
func (p Path) Invoke(args ...core.Any) (core.Any, error) {
	if len(args) == 0 {
		// Default behavior: return the node as File/Directory/Node
		return p.get(context.Background())
	}

	// First argument should be a keyword
	method, ok := args[0].(builtin.Keyword)
	if !ok {
		return nil, fmt.Errorf("path method must be a keyword, got %T", args[0])
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	switch string(method) {
	case ":cat":
		return p.cat(ctx)
	default:
		return nil, fmt.Errorf("unknown path method: %s (supported: :cat)", method)
	}
}

// cat returns the content of an IPFS file as []byte
func (p Path) cat(ctx context.Context) (core.Any, error) {
	if p.Env == nil || p.Env.IPFS == nil {
		return nil, fmt.Errorf("IPFS client not initialized")
	}

	// Get the node from IPFS
	node, err := p.Env.IPFS.Unixfs().Get(ctx, p.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPFS path: %w", err)
	}

	// Handle different node types
	switch node := node.(type) {
	case files.File:
		// For files, read the content into a byte slice
		content, err := io.ReadAll(node)
		if err != nil {
			return nil, fmt.Errorf("failed to read file content: %w", err)
		}
		return content, nil
	case files.Directory:
		return nil, fmt.Errorf("path is a directory, use :get instead of :cat")
	default:
		return nil, fmt.Errorf("unexpected node type: %T", node)
	}
}

// get returns the IPFS node as a file-like object (io.Reader) that can be used with eval
func (p Path) get(ctx context.Context) (core.Any, error) {
	if p.Env == nil || p.Env.IPFS == nil {
		return nil, fmt.Errorf("IPFS client not initialized")
	}

	// Get the node from IPFS
	node, err := p.Env.IPFS.Unixfs().Get(ctx, p.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPFS path: %w", err)
	}

	// Return the node as a file-like object
	switch node := node.(type) {
	case files.File:
		return File{File: node}, nil
	case files.Directory:
		return Directory{Directory: node}, nil
	default:
		return Node{Node: node}, nil
	}
}
