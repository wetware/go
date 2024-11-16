package call

import (
	"bytes"
	"io"
	"os"

	"capnproto.org/go/capnp/v3"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/proc"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "call",
		Usage:     "generate method call data",
		ArgsUsage: "<method>",
		Flags: []cli.Flag{
			&cli.Uint64SliceFlag{
				Name:  "push",
				Usage: "push a word onto the guest stack",
			},
		},
		Action: func(c *cli.Context) error {
			m, seg := capnp.NewSingleSegmentMessage(nil)
			defer m.Release()

			call, err := proc.NewRootMethodCall(seg)
			if err != nil {
				return err
			}

			if err := call.SetName(c.Args().First()); err != nil {
				return err
			}

			stack := c.Uint64Slice("push")
			size := int32(len(stack))

			s, err := call.NewStack(size)
			if err != nil {
				return err
			}
			for i, word := range stack {
				s.Set(i, word)
			}

			if b, err := message(c); err != nil {
				return err
			} else if err = call.SetCallData(b); err != nil {
				return err
			}

			if b, err := m.Marshal(); err != nil {
				return err
			} else if _, err = io.Copy(os.Stdout, bytes.NewReader(b)); err != nil {
				return err
			}

			return nil
		},
	}
}

func message(c *cli.Context) ([]byte, error) {
	switch f := c.App.Reader.(type) {
	case *os.File:
		info, err := f.Stat()
		if err != nil {
			return nil, err
		}

		if info.Size() > 0 {
			return io.ReadAll(f)
		}
	}

	return nil, nil
}
