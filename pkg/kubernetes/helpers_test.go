package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelpers_trimQuotes(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		data := []byte{}
		require.Equal(t, "", string(trimQuotes(data)))
	})

	t.Run("only quotes", func(t *testing.T) {
		data := []byte{'"', '"'}
		require.Equal(t, "", string(trimQuotes(data)))
	})

	t.Run("double-quoted data", func(t *testing.T) {
		data := []byte{'"', 'p', 'o', 'r', 't', 'e', 'r', '"'}
		require.Equal(t, "porter", string(trimQuotes(data)))
	})

	t.Run("single-quoted data", func(t *testing.T) {
		data := []byte{'\'', 'p', 'o', 'r', 't', 'e', 'r', '\''}
		require.Equal(t, "porter", string(trimQuotes(data)))
	})

	t.Run("only beginning quote", func(t *testing.T) {
		data := []byte{'\'', 'p', 'o', 'r', 't', 'e', 'r'}
		require.Equal(t, "'porter", string(trimQuotes(data)))
	})

	t.Run("no quotes", func(t *testing.T) {
		data := []byte{'p', 'o', 'r', 't', 'e', 'r'}
		require.Equal(t, "porter", string(trimQuotes(data)))
	})
}
