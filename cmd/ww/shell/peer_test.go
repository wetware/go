package shell_test

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p"
	"github.com/spy16/slurp/builtin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wetware/go/cmd/ww/shell"
)

func TestPeer(t *testing.T) {
	t.Parallel()

	// Create a test host
	host, err := libp2p.New()
	require.NoError(t, err, "failed to create test host")
	defer host.Close()

	// Create a test peer
	peer := &shell.Peer{
		Ctx:  context.Background(),
		Host: host,
	}

	t.Run("String", func(t *testing.T) {
		str := peer.String()
		assert.NotEqual(t, "<Peer: (no host)>", str, "expected peer string to show host ID")
		assert.NotEmpty(t, str, "expected non-empty peer string")
	})

	t.Run("ID", func(t *testing.T) {
		id := peer.ID()
		require.NotNil(t, id, "expected non-nil peer ID")

		idStr, ok := id.(builtin.String)
		require.True(t, ok, "expected builtin.String, got %T", id)
		assert.NotEmpty(t, string(idStr), "expected non-empty peer ID string")
	})

	t.Run("IsSelf", func(t *testing.T) {
		// Test with our own peer ID
		ourID := host.ID().String()
		result, err := peer.IsSelf(ourID)
		require.NoError(t, err, "unexpected error")

		isSelf, ok := result.(builtin.Bool)
		require.True(t, ok, "expected builtin.Bool, got %T", result)
		assert.True(t, bool(isSelf), "expected peer to recognize itself")

		// Test with a different peer ID - create another host to get a valid peer ID
		otherHost, err := libp2p.New()
		require.NoError(t, err, "failed to create other host")
		defer otherHost.Close()

		differentID := otherHost.ID().String()
		result, err = peer.IsSelf(differentID)
		require.NoError(t, err, "unexpected error")

		isSelf, ok = result.(builtin.Bool)
		require.True(t, ok, "expected builtin.Bool, got %T", result)
		assert.False(t, bool(isSelf), "expected peer to not recognize different ID as self")
	})

	t.Run("Invoke", func(t *testing.T) {
		// Test :id method
		result, err := peer.Invoke(builtin.Keyword("id"))
		require.NoError(t, err, "unexpected error calling :id")
		assert.NotNil(t, result, "expected non-nil result from :id")

		// Test :is-self method
		ourID := host.ID().String()
		result, err = peer.Invoke(builtin.Keyword("is-self"), ourID)
		require.NoError(t, err, "unexpected error calling :is-self")
		assert.NotNil(t, result, "expected non-nil result from :is-self")

		// Test invalid method
		_, err = peer.Invoke(builtin.Keyword("invalid"))
		assert.Error(t, err, "expected error for invalid method")
	})

	t.Run("BuiltinStringSupport", func(t *testing.T) {
		// Test that builtin.String types are properly handled in Peer methods.
		// This is important for composable expressions like:
		//   (peer :send (peer :id) "/ww/0.1.0/crU9ZDuzKWr" "hello, wetware!")
		// where (peer :id) returns a builtin.String, not a regular Go string.
		ourID := host.ID().String()
		builtinID := builtin.String(ourID)

		// Test :is-self with builtin.String
		result, err := peer.Invoke(builtin.Keyword("is-self"), builtinID)
		require.NoError(t, err, "unexpected error calling :is-self with builtin.String")
		assert.NotNil(t, result, "expected non-nil result from :is-self")

		// Test :connect with builtin.String (this will fail to connect, but should not error on type)
		_, err = peer.Invoke(builtin.Keyword("connect"), builtinID)
		// We expect a connection error, not a type error
		assert.Error(t, err, "expected connection error")
		assert.NotContains(t, err.Error(), "must be a string", "should not get type error for builtin.String")
	})
}
