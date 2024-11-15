package system

import "io"

type IO struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Args, Env      []string
}
