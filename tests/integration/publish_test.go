//go:build integration
// +build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

func TestPublish_BuildWithVersionOverride(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	// Create a bundle
	err := p.Create()
	require.NoError(t, err)

	// Build with version override
	buildOpts := porter.BuildOptions{}
	buildOpts.Version = "0.0.0"

	err = buildOpts.Validate(p.Porter)
	require.NoError(t, err)

	err = p.Build(ctx, buildOpts)
	require.NoError(t, err)

	publishOpts := porter.PublishOptions{}
	publishOpts.Registry = "localhost:5000"
	err = publishOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for bundle failed")

	// Confirm that publish picks up the version override
	// (Otherwise, image tagging and publish will fail)
	err = p.Publish(ctx, publishOpts)
	require.NoError(p.T(), err, "publish of bundle failed")
}
