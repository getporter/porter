package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCachedBundle_GetBundleID(t *testing.T) {
	cb := CachedBundle{}

	cb.Tag = kahn1dot01
	bid := cb.GetBundleID()
	require.Equal(t, kahn1dot0Hash, bid, "hashing the bundle ID twice should be the same")

	cb.Tag = kahnlatest
	bid2 := cb.GetBundleID()
	assert.NotEqual(t, kahn1dot0Hash, bid2, "different tags should result in different hashes")
}
