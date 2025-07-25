package system

import (
	"context"
	"errors"
)

type ExecConfig struct{}

func (e ExecConfig) New() Executor {
	return Executor_ServerToClient(e)
}

func (e ExecConfig) Spawn(ctx context.Context, call Executor_spawn) error {
	return errors.New("ExecConfig::Spawn:  NOT IMPLEMENTED")
}
