// //go:generate tinygo build -o main.wasm -target=wasi -scheduler=none main.go

package main

// import (
// 	"bytes"
// 	"context"
// 	"io"
// 	"log/slog"
// 	"os"

// 	"capnproto.org/go/capnp/v3"
// 	"github.com/urfave/cli/v2"
// 	"github.com/wetware/go/proc"
// )

// func main() {
// 	ctx := context.TODO()

// 	app := &cli.App{
// 		Name:      "deliver",
// 		Usage:     "read message from stdn and send to `PID`",
// 		ArgsUsage: "<PID>",
// 		Flags: []cli.Flag{
// 			&cli.StringFlag{
// 				Name:    "method",
// 				Aliases: []string{"m"},
// 				Usage:   "method name",
// 			},
// 			&cli.Uint64SliceFlag{
// 				Name:    "push",
// 				Aliases: []string{"p"},
// 				Usage:   "push u64 onto stack",
// 			},
// 		},
// 		Action: deliver,
// 	}

// 	if err := app.RunContext(ctx, os.Args); err != nil {
// 		slog.ErrorContext(ctx, "application failed",
// 			"reason", err)
// 		os.Exit(1)
// 	}
// }

// func deliver(c *cli.Context) error {
// 	name := c.Args().First()
// 	f, err := os.Open(name)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	m, seg := capnp.NewSingleSegmentMessage(nil)
// 	defer m.Release()

// 	call, err := proc.NewRootMethodCall(seg)
// 	if err != nil {
// 		return err
// 	}

// 	err = call.SetName(c.String("method"))
// 	if err != nil {
// 		return err
// 	}

// 	stack := c.Uint64Slice("stack")
// 	size := int32(len(stack))

// 	callStack, err := call.NewStack(size)
// 	if err != nil {
// 		return err
// 	}

// 	for i, word := range stack {
// 		callStack.Set(i, word)
// 	}

// 	r := io.LimitReader(c.App.Reader, 1<<32-1) // max u32
// 	data, err := io.ReadAll(r)
// 	if err != nil {
// 		return err
// 	}

// 	if err = call.SetCallData(data); err != nil {
// 		return err
// 	}

// 	b, err := m.Marshal()
// 	if err != nil {
// 		return err
// 	}

// 	n, err := io.Copy(f, bytes.NewReader(b))
// 	if err != nil {
// 		return err
// 	}

// 	slog.DebugContext(c.Context, "delivered message",
// 		"size", n,
// 		"dest", name)
// 	return nil
// }
