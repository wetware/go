package shell

import (
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/spy16/slurp"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/urfave/cli/v2"
)

const helpMessage = `Wetware Shell - Available commands:
help                    - Show this help message
version                 - Show wetware version
(+ a b ...)            - Sum numbers
(* a b ...)            - Multiply numbers
(= a b)                - Compare equality
(> a b)                - Greater than
(< a b)                - Less than
(println expr)         - Print expression with newline
(print expr)           - Print expression without newline
(import "module")      - Import a module (stubbed)

IPFS Path Syntax:
/ipfs/QmHash/...       - Direct IPFS path
/ipns/domain/...       - IPNS path

P2P Commands (use --with-p2p):
(peer :send "peer-addr" "proc-id" data) - Send data to a peer process
(peer :connect "peer-addr") - Connect to a peer
(peer :is-self "peer-id") - Check if peer ID is our own
(peer :id)             - Get our own peer ID`

var globals = map[string]core.Any{
	// Basic values
	"nil":     builtin.Nil{},
	"true":    builtin.Bool(true),
	"false":   builtin.Bool(false),
	"version": builtin.String("wetware-0.1.0"),

	// Basic operations
	"=": slurp.Func("=", core.Eq),
	"+": slurp.Func("sum", func(a ...int) int {
		sum := 0
		for _, item := range a {
			sum += item
		}
		return sum
	}),
	">": slurp.Func(">", func(a, b builtin.Int64) bool {
		return a > b
	}),
	"<": slurp.Func("<", func(a, b builtin.Int64) bool {
		return a < b
	}),
	"*": slurp.Func("*", func(a ...int) int {
		product := 1
		for _, item := range a {
			product *= item
		}
		return product
	}),
	"/": slurp.Func("/", func(a, b builtin.Int64) float64 {
		return float64(a) / float64(b)
	}),

	// Wetware-specific functions
	"help": slurp.Func("help", func() string {
		return helpMessage
	}),
	"println": slurp.Func("println", func(args ...core.Any) {
		for _, arg := range args {
			fmt.Println(arg)
		}
	}),
	"print": slurp.Func("print", func(args ...core.Any) {
		for _, arg := range args {
			fmt.Print(arg)
		}
	}),
}

func NewGlobals(c *cli.Context) (map[string]core.Any, error) {
	gs := make(map[string]core.Any, len(globals))
	for k, v := range globals {
		gs[k] = v
	}

	// Add IPFS support if --with-ipfs flag is set
	if c.Bool("with-ipfs") || c.Bool("with-all") {
		if env.IPFS == nil {
			return nil, errors.New("uninitialized IPFS environment")
		}
		gs["ipfs"] = &IPFS{CoreAPI: env.IPFS}
	}

	// Add P2P functionality if --with-p2p flag is set
	if c.Bool("with-p2p") || c.Bool("with-all") {
		// Create a new host for P2P functionality
		host, err := libp2p.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create libp2p host: %v", err)
		}
		gs["peer"] = &Peer{
			Ctx:  c.Context,
			Host: host,
		}
	}

	return gs, nil
}
