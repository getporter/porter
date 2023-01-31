package plugins

import (
	"testing"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallOptions_Validate(t *testing.T) {
	// InstallOptions is already tested in pkgmgmt, we just want to make sure DefaultFeedURL is set
	cxt := portercontext.NewTestContext(t)
	opts := InstallOptions{}
	err := opts.Validate([]string{"pkg1"}, cxt.Context)
	require.NoError(t, err, "Validate failed")
	assert.NotEmpty(t, opts.FeedURL, "Feed URL was not defaulted to the plugins feed URL")
}

func TestInstallPluginsConfig(t *testing.T) {
	input := InstallPluginsConfig{"kubernetes": pkgmgmt.InstallOptions{URL: "test-kubernetes.com"}, "azure": pkgmgmt.InstallOptions{URL: "test-azure.com"}}
	expected := []pkgmgmt.InstallOptions{{Name: "azure", PackageType: "plugin", URL: "test-azure.com"}, {Name: "kubernetes", PackageType: "plugin", URL: "test-kubernetes.com"}}

	cfg := NewInstallPluginConfigs(input)
	require.Equal(t, expected, cfg.Values())
}
