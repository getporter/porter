package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCachedBundle_GetBundleID(t *testing.T) {
	t.Parallel()

	cb := CachedBundle{}

	cb.Reference = kahn1dot01
	bid := cb.GetBundleID()
	assert.Equal(t, kahn1dot0Hash, bid, "hashing the bundle ID twice should be the same")

	cb.Reference = kahnlatest
	bid2 := cb.GetBundleID()
	assert.NotEqual(t, kahn1dot0Hash, bid2, "different tags should result in different hashes")
}
