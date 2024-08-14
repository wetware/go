package system_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wetware/go/system"
)

func TestMemorySegment(t *testing.T) {
	t.Parallel()

	seg := system.NewMemorySegment(1, 2)
	t.Logf("%064b", seg)

	assert.Equal(t, uint32(1), seg.Offset())
	assert.Equal(t, uint32(2), seg.Length())
}
