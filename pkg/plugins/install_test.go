package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallOptions_Validate(t *testing.T) {
	// InstallOptions is already tested in pkgmgmt, we just want to make sure DefaultFeedURL is set
	opts := InstallOptions{}
	err := opts.Validate([]string{"pkg1"})
	require.NoError(t, err, "Validate failed")
	assert.Equal(t, DefaultFeedUrl, opts.FeedURL, "Feed URL was not defaulted to the plugins feed URL")
}
