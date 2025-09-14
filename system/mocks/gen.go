//go:generate mockgen -source=gen.go -destination=mock_stream.go -package=mocks

package mocks

import (
	"github.com/libp2p/go-libp2p/core/network"
)

// StreamInterface embeds network.Stream to provide a clean interface for mocking
type StreamInterface interface {
	network.Stream
}
