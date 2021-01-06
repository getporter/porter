// +build integration

package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"get.porter.sh/porter/pkg/porter"
)

func TestArchive(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	bundleName := p.AddTestBundleDir("../build/testdata/bundles/mysql", true)
	tag := fmt.Sprintf("localhost:5000/%s:v0.1.3", bundleName)

	// Currently, archive requires the bundle to already be published.
	// https://github.com/getporter/porter/issues/697
	publishOpts := porter.PublishOptions{}
	publishOpts.Tag = tag
	err := publishOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for bundle failed")

	err = p.Publish(publishOpts)
	require.NoError(p.T(), err, "publish of bundle failed")

	// Archive bundle
	archiveOpts := porter.ArchiveOptions{}
	archiveOpts.Tag = tag
	err = archiveOpts.Validate([]string{"mybuns.tgz"}, p.Porter)
	require.NoError(p.T(), err, "validation of archive opts for bundle failed")

	err = p.Archive(archiveOpts)
	require.NoError(p.T(), err, "archival of bundle failed")

	info, err := p.FileSystem.Stat("mybuns.tgz")
	require.NoError(p.T(), err)
	require.Equal(p.T(), os.FileMode(0644), info.Mode())

	// Publish bundle from archive, with new tag
	publishFromArchiveOpts := porter.PublishOptions{
		ArchiveFile: "mybuns.tgz",
		BundlePullOptions: porter.BundlePullOptions{
			Tag: fmt.Sprintf("localhost:5000/archived-%s:v0.1.3", bundleName),
		},
	}
	err = publishFromArchiveOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for bundle failed")

	err = p.Publish(publishFromArchiveOpts)
	require.NoError(p.T(), err, "publish of bundle from archive failed")
}
