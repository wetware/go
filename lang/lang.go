package lang

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/system"
)

func NewExecutor(ctx context.Context, client system.Executor) core.Any {
	return &Executor{Client: client}
}

var _ core.Invokable = (*Executor)(nil)

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
