package util

import (
	"context"
	"log/slog"

	"github.com/thejerf/suture/v4"
)

var EventHook = EventHookWithContext(context.Background())

func EventHookWithContext(ctx context.Context) suture.EventHook {
	return func(e suture.Event) {
		switch e.Type() {
		case suture.EventTypeBackoff:
			ev := e.(suture.EventBackoff)
			slog.InfoContext(ctx, "entered backoff state",
				"supervisor", ev.SupervisorName)

		case suture.EventTypeResume:
			ev := e.(suture.EventResume)
			slog.InfoContext(ctx, "resumed",
				"supervisor", ev.SupervisorName)

		case suture.EventTypeServicePanic:
			ev := e.(suture.EventServicePanic)
			slog.ErrorContext(ctx, "service panicked",
				"supervisor", ev.SupervisorName,
				"service", ev.ServiceName,
				"reason", ev.PanicMsg)

		case suture.EventTypeServiceTerminate:
			ev := e.(suture.EventServiceTerminate)
			slog.WarnContext(ctx, "service terminated",
				"supervisor", ev.SupervisorName,
				"service", ev.ServiceName,
				"reason", ev.Err)

		case suture.EventTypeStopTimeout:
			ev := e.(suture.EventStopTimeout)
			slog.ErrorContext(ctx, "service failed to stop",
				"supervisor", ev.SupervisorName,
				"service", ev.ServiceName)

		default:
			panic(e) // unhandled event
		}
	}
}
