package system

import "io"

type Cmd struct {
	Path           string
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Args, Env      []string
}
