package deliver

import (
	"fmt"
	"io"
	"path"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/urfave/cli/v2"
	ww "github.com/wetware/go"
	"github.com/wetware/go/proc"
	"github.com/wetware/go/system"
)

var args Args

func Command() *cli.Command {
	return &cli.Command{
		Name: "deliver",
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:  "push",
				Usage: "push a 64-bit word onto the call stack",
			},
		},
		Before: args.Bind,
		Action: deliver,
	}
}

func deliver(c *cli.Context) error {
	// 1. Set up libp2p and peer discovery
	////
	h, err := ww.NewP2PHostWithMDNS(c.Context)
	if err != nil {
		return err
	}
	defer h.Close()

	// 2. Open a stream to peer
	////
	s, err := h.NewStream(c.Context, args.Peer, args.Protocol())
	if err != nil {
		return err
	}
	defer s.Close()

	// 3. Round trip
	////
	return roundTrip(c, s)
}

func roundTrip(c *cli.Context, s network.Stream) error {
	if n, err := io.Copy(s, c.App.Reader); err != nil { // Request
		return fmt.Errorf("wrote %d bytes: %w", n, err)
	} else if n, err := io.Copy(c.App.Writer, s); err != nil { // Response
		return fmt.Errorf("read %d bytes: %w", n, err)
	}

	return nil
}

type Args struct {
	Peer peer.ID
	PID  proc.PID
	Call proc.Call
}

func (a *Args) Bind(c *cli.Context) (err error) {
	arg0 := c.Args().Get(0)
	arg1 := c.Args().Get(1)

	a.Call = proc.Call{
		Method: c.Args().Get(2),
		Stack:  c.Uint64Slice("push"),
	}

	if a.Peer, err = peer.Decode(arg0); err != nil {
		err = fmt.Errorf("arg[0]: decode peer id: %w", err)
	} else if a.PID, err = proc.ParsePID(arg1); err != nil {
		err = fmt.Errorf("arg[1]: parse pid: %w", err)
	}

	return
}

func (a Args) Protocol() protocol.ID {
	p2p := "/p2p/" + a.Peer.String()
	pid := "/pid/" + a.PID.String()
	proto := path.Join(p2p, system.Proto.String(), pid)
	return protocol.ID(proto)
}
