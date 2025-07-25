package lang

import (
	"context"
	"fmt"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/system"
)

// ConsolePrintln implements a standalone println function for the shell
type ConsolePrintln struct {
	Console system.Console
}

// Invoke implements core.Invokable for ConsolePrintln
func (cp ConsolePrintln) Invoke(args ...core.Any) (core.Any, error) {
	// Identity law: when called with no arguments, return self
	if len(args) == 0 {
		return cp, nil
	}

	if len(args) != 1 {
		return nil, fmt.Errorf("println requires exactly 1 argument, got %d", len(args))
	}

	// Extract the data to print from the argument
	var data string
	switch arg := args[0].(type) {
	case string:
		data = arg
	case builtin.String:
		data = string(arg)
	case *Buffer:
		data = arg.String()
	default:
		data = fmt.Sprintf("%v", arg)
	}

	// Call the Console.Println method
	ctx := context.Background()
	future, release := cp.Console.Println(ctx, func(call system.Console_println_Params) error {
		return call.SetOutput(data)
	})
	defer release()

	res, err := future.Struct()
	if err != nil {
		return nil, fmt.Errorf("failed to get println results: %w", err)
	}

	// Return the number of bytes written
	return builtin.Int64(res.N()), nil
}
