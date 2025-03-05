package syncutils_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	syncutils "github.com/wetware/go/util/sync"
)

func TestAny(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		var any syncutils.Any

		any.Go(func() error {
			return nil
		})

		assert.NoError(t, any.Wait())
	})

	t.Run("Error", func(t *testing.T) {
		var any syncutils.Any
		expected := errors.New("test error")

		any.Go(func() error {
			return expected
		})

		assert.ErrorIs(t, any.Wait(), expected)
	})

	t.Run("FirstError", func(t *testing.T) {
		var any syncutils.Any
		first := errors.New("first error")
		second := errors.New("second error")

		any.Go(func() error {
			return first
		})

		any.Go(func() error {
			return second
		})

		assert.Contains(t, []error{first, second}, any.Wait(),
			"unexpected error: %v", any.Wait())
	})
}
