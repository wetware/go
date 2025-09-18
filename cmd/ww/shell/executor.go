package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/libp2p/go-libp2p/core/protocol"
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

		protocol, err := e.ExecBytes(ctx, bytecode)
		if err != nil {
			return nil, fmt.Errorf("failed to execute bytecode: %w", err)
		}
		return protocol, nil

	case files.Directory:
		return nil, errors.New("TODO:  directory support")
	default:
		return nil, fmt.Errorf("unexpected node type: %T", node)
	}
}

func (e Exec) ExecBytes(ctx context.Context, bytecode []byte) (protocol.ID, error) {
	f, release := e.Session.Exec().Exec(ctx, func(p system.Executor_exec_Params) error {
		return p.SetBytecode(bytecode)
	})
	defer release()

	// Wait for the protocol setup to complete
	result, err := f.Struct()
	if err != nil {
		return "", err
	}

	proto, err := result.Protocol()
	return protocol.ID(proto), err
}

func (e Exec) NewContext(opts map[builtin.Keyword]core.Any) (context.Context, context.CancelFunc) {
	// TODO:  add support for parsing durations like 15s, 15m, 15h, 15d
	// if timeout, ok := opts["timeout"].(time.Duration); ok {
	// 	return context.WithTimeout(context.Background(), timeout)
	// }

	return context.WithTimeout(context.Background(), time.Second*15)
}
