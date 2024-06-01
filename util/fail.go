package util

import (
	"fmt"

	"capnproto.org/go/capnp/v3"
)

func Fail[T ~capnp.ClientKind](err error) T {
	client := capnp.ErrorClient(err)
	return T(client)
}

func Failf[T ~capnp.ClientKind](format string, args ...any) T {
	return Fail[T](fmt.Errorf(format, args...))
}
