package boot

import "context"

type Bootstrapper interface {
	// Bootstrap allows callers to hint to the routing system to get into a
	// Boostrapped state and remain there.
	Bootstrap(context.Context) error
}
