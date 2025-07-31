package system

import (
	"context"
	"errors"
	"os/exec"
)

var _ Executor_Server = (*ExecConfig)(nil)

type ExecConfig struct{ Enabled bool }

func (e ExecConfig) New() Executor {
	return Executor_ServerToClient(e)
}

func (e ExecConfig) Spawn(ctx context.Context, call Executor_spawn) error {
	if !e.Enabled {
		return errors.New("executor is disabled")
	}

	ipfs := call.Args().IPFS()
	defer ipfs.Release()

	params, err := call.Args().Command()
	if err != nil {
		return err
	}

	path, err := params.Path()
	if err != nil {
		return err
	}

	stat, release := ipfs.Stat(ctx, func(p IPFS_stat_Params) error {
		return p.SetCid(path)
	})
	defer release()

	info, err := stat.Info().Struct()
	if err != nil {
		return err
	}

	args, err := params.Args()
	if err != nil {
		return err
	}

	// Check if the node is a file
	if info.NodeType().Which() != NodeInfo_nodeType_Which_file {
		return errors.New("path is not a file")
	}

	// Convert capnp.TextList to []string
	argsList := make([]string, args.Len())
	for i := 0; i < args.Len(); i++ {
		arg, err := args.At(i)
		if err != nil {
			return err
		}
		argsList[i] = arg
	}

	cmd := exec.CommandContext(ctx, path, argsList...)
	return cmd.Start()
}

var _ Cell_Server = (*CellConfig)(nil)

type CellConfig struct {
	Cmd *exec.Cmd
}

func (c CellConfig) Wait(ctx context.Context, call Cell_wait) error {
	return c.Cmd.Wait()
}
