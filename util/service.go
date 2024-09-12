package util

import (
	"context"
	"log/slog"

	"github.com/thejerf/suture/v4"
)

var EventHook = EventHookWithContext(context.Background())

func EventHookWithContext(ctx context.Context) suture.EventHook {
	return func(e suture.Event) {
		var args []any
		for k, v := range e.Map() {
			args = append(args, k, v)
		}

		switch e.Type() {
		case suture.EventTypeBackoff:
			slog.InfoContext(ctx, e.String(), args...)

		case suture.EventTypeResume:
			slog.InfoContext(ctx, e.String(), args...)

		case suture.EventTypeServicePanic:
			slog.WarnContext(ctx, e.String(), args...)

		case suture.EventTypeServiceTerminate:
			slog.WarnContext(ctx, e.String(), args...)

		case suture.EventTypeStopTimeout:
			slog.ErrorContext(ctx, e.String(), args...)

		default:
			panic(e) // unhandled event
		}
	}
}
