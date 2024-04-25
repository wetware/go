package ww

import (
	"fmt"
	"reflect"
)

type UnhandledEvent struct {
	Event any
}

func (err UnhandledEvent) Error() string {
	return fmt.Sprintf("unhandled event: %s", reflect.TypeOf(err.Event))
}
