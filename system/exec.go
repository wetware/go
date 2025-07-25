package system

import (
	"context"
	"errors"
	"os/exec"
)

var _ Executor_Server = (*ExecConfig)(nil)

type ExecConfig struct {
	Enabled bool
	IPFS    IPFS
}

func (e ExecConfig) New() Executor {
	return Executor_ServerToClient(e)
}

func (e ExecConfig) Spawn(ctx context.Context, call Executor_spawn) error {
	// ...

	return errors.New("ExecConfig::Spawn:  NOT IMPLEMENTED")
}

var _ Cell_Server = (*CellConfig)(nil)

type CellConfig struct {
	Cmd *exec.Cmd
}

func (c CellConfig) Wait(ctx context.Context, call Cell_wait) error {
	return c.Cmd.Wait()
}
