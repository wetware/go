package system_test

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/wetware/go/system"
)

func TestNet_MatchProto(t *testing.T) {
	t.Parallel()

	n := system.Net{}
	t.Log(system.Proto.Unwrap())

	for _, tt := range []struct {
		proto protocol.ID
		fail  bool
	}{
		{
			proto: system.Proto.Unwrap() + "/3xRb2SHLJR34eqxuzaQgEx6AwmWG",
		},
		{
			proto: system.Proto.Unwrap(),
			fail:  true,
		},
	} {
		if matched := n.MatchProto(tt.proto); tt.fail {
			assert.False(t, matched, "should not match %s", tt.proto)
		} else {
			assert.True(t, matched, "should match %s", tt.proto)
		}
	}
}
