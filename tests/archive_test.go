// +build integration

package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/deislabs/porter/pkg/porter"
)

func TestArchive(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	p.TestConfig.TestContext.AddTestFile(filepath.Join(p.TestDir, "testdata/bundles/porter-busybox/porter.yaml"), "porter.yaml")
	p.TestConfig.TestContext.AddTestFile(filepath.Join(p.TestDir, "testdata/bundles/porter-busybox/Dockerfile.tmpl"), "Dockerfile.tmpl")

	// Currently, archive requires the bundle to already be published.
	// https://github.com/deislabs/porter/issues/697
	publishOpts := porter.PublishOptions{}
	err := publishOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for bundle failed")

	err = p.Publish(publishOpts)
	require.NoError(p.T(), err, "publish of bundle failed")

	// Archive bundle

	archiveOpts := porter.ArchiveOptions{}
	err = archiveOpts.Validate([]string{"mybuns.tgz"}, p.Context)
	require.NoError(p.T(), err, "validation of archive opts for bundle failed")

	err = p.Archive(archiveOpts)
	require.NoError(p.T(), err, "archival of bundle failed")

	// Publish bundle from archive, with new tag

	publishFromArchiveOpts := porter.PublishOptions{
		ArchiveFile: "mybuns.tgz",
		BundlePullOptions: porter.BundlePullOptions{
			Tag: "deislabs/porter-busybox-from-archive:v0.1.0",
		},
	}
	err = publishFromArchiveOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for bundle failed")

	err = p.Publish(publishFromArchiveOpts)
	require.NoError(p.T(), err, "publish of bundle from archive failed")
}
