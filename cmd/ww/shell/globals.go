package shell

import (
	"fmt"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
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
(send "peer-addr-or-id" "proc-id" data) - Send data to a peer process (data: string, []byte, or io.Reader)
(import "module")      - Import a module (stubbed)

IPFS Path Syntax:
/ipfs/QmHash/...       - Direct IPFS path
/ipns/domain/...       - IPNS path`

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
	"send": slurp.Func("send", func(peerAddr, procId string, data interface{}) error {
		return SendToPeer(peerAddr, procId, data)
	}),
}
