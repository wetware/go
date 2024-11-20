package system

import "io"

type Cmd struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Args, Env      []string
}
