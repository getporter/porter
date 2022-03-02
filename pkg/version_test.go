package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserAgent(t *testing.T) {
	t.Run("append versions when available", func(t *testing.T) {
		Version = "v1.0.0"
		Commit = "abc123"

		require.Contains(t, UserAgent(), PORTER_USER_AGENT+"/"+Version)
	})

	t.Run("append commit hash when version is not available", func(t *testing.T) {
		Version = ""
		Commit = "abc123"

		require.Contains(t, UserAgent(), PORTER_USER_AGENT+"/"+Commit)
	})

	t.Run("omit slash when neither version nor commit hash is available", func(t *testing.T) {
		Version = ""
		Commit = ""

		require.Contains(t, UserAgent(), PORTER_USER_AGENT)
	})
}
