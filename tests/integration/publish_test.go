// +build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"

	"get.porter.sh/porter/pkg/porter"
)

func TestPublish_BuildWithVersionOverride(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	// Create a bundle
	err := p.Create()
	require.NoError(t, err)

	// Build with version override
	buildOpts := porter.BuildOptions{}
	buildOpts.Version = "0.0.0"

	err = buildOpts.Validate(p.Context)
	require.NoError(t, err)

	err = p.Build(buildOpts)
	require.NoError(t, err)

	publishOpts := porter.PublishOptions{}
	publishOpts.Registry = "localhost:5000"
	err = publishOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for bundle failed")

	// Confirm that publish picks up the version override
	// (Otherwise, image tagging and publish will fail)
	err = p.Publish(publishOpts)
	require.NoError(p.T(), err, "publish of bundle failed")
}
