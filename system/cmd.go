package system

import "io"

type Cmd struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Args, Env      []string
}

func (cmd Cmd) ExecPath() string {
	if len(cmd.Args) == 0 {
		return ""
	}

	return cmd.Args[0]
}

func (cmd Cmd) Arguments() []string {
	if len(cmd.Args) < 2 {
		return nil
	}

	return cmd.Args[1:]
}
