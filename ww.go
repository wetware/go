//go:generate capnp compile -I$GOPATH/src/capnproto.org/go/capnp/std -ogo ww.capnp

package ww

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/std/capnp/schema"
)

type Node[T ~capnp.ClientKind] interface {
	Client() T
	Type() schema.Node_Future
}

func Connect[T ~capnp.ClientKind](n Node[T]) T {
	// TODO:  some kind of schema validation check should go here,
	//        and we should return an ErrorClient if the client is
	//        not what we expect (per schema).

	return T(n.Client()) // succeed
}
