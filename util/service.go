package util

import (
	"log/slog"

	"github.com/thejerf/suture/v4"
)

func EventHook(e suture.Event) {
	var args []any
	for k, v := range e.Map() {
		args = append(args, k, v)
	}

	switch e.Type() {
	case suture.EventTypeBackoff:
		slog.Info(e.String(), args...)

	case suture.EventTypeResume:
		slog.Info(e.String(), args...)

	case suture.EventTypeServicePanic:
		slog.Warn(e.String(), args...)

	case suture.EventTypeServiceTerminate:
		slog.Warn(e.String(), args...)

	case suture.EventTypeStopTimeout:
		slog.Error(e.String(), args...)

	default:
		panic(e) // unhandled event
	}
}
