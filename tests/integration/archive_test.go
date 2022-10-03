//go:build integration

package integration

import (
	"crypto/sha256"
	"fmt"
	"io"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Validate that archiving a bundle twice results in the same digest
func TestArchive_StableDigest(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	// Use a fixed bundle to work with so that we can rely on the registry and layer digests
	const reference = "ghcr.io/getporter/examples/whalegap:v0.2.0"

	// Archive bundle
	archive1Opts := porter.ArchiveOptions{}
	archive1Opts.Reference = reference
	archiveFile1 := "mybuns1.tgz"
	err := archive1Opts.Validate(ctx, []string{archiveFile1}, p.Porter)
	require.NoError(p.T(), err, "validation of archive opts for bundle failed")

	err = p.Archive(ctx, archive1Opts)
	require.NoError(p.T(), err, "archival of bundle failed")

	info, err := p.FileSystem.Stat(archiveFile1)
	require.NoError(p.T(), err)
	tests.AssertFilePermissionsEqual(t, archiveFile1, pkg.FileModeWritable, info.Mode())

	hash1 := getHash(p, archiveFile1)

	// Check to be sure the shasum is stable after archiving a second time
	archive2Opts := porter.ArchiveOptions{}
	archive2Opts.Reference = reference
	archiveFile2 := "mybuns2.tgz"
	err = archive2Opts.Validate(ctx, []string{archiveFile2}, p.Porter)
	require.NoError(p.T(), err, "validation of archive opts for bundle failed")

	err = archive1Opts.Validate(ctx, []string{archiveFile2}, p.Porter)
	require.NoError(t, err, "Second validate failed")

	err = p.Archive(ctx, archive2Opts)
	require.NoError(t, err, "Second archive failed")
	assert.Equal(p.T(), hash1, getHash(p, archiveFile2), "shasum of archive did not stay the same on the second call to archive")
	// Publish bundle from archive, with new reference
	localReference := "localhost:5000/archived-whalegap:v0.2.0"
	publishFromArchiveOpts := porter.PublishOptions{
		ArchiveFile: archiveFile1,
		BundlePullOptions: porter.BundlePullOptions{
			Reference: localReference,
		},
	}
	err = publishFromArchiveOpts.Validate(p.Config)
	require.NoError(p.T(), err, "validation of publish opts for bundle failed")

	err = p.Publish(ctx, publishFromArchiveOpts)
	require.NoError(p.T(), err, "publish of bundle from archive failed")

	// Archive from the newly published bundle in local registry
	archive3Opts := porter.ArchiveOptions{}
	archive3Opts.Reference = localReference
	archiveFile3 := "mybuns3.tgz"
	err = archive3Opts.Validate(ctx, []string{archiveFile3}, p.Porter)
	require.NoError(p.T(), err, "validation of archive opts for bundle failed")
	err = p.Archive(ctx, archive3Opts)
	require.NoError(t, err, "archive from the published bundle in local registry failed")
}

func getHash(p *porter.TestPorter, path string) string {
	f, err := p.FileSystem.Open(path)
	require.NoError(p.T(), err, "opening archive failed")
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	require.NoError(p.T(), err, "hashing of archive failed")

	return fmt.Sprintf("%x", h.Sum(nil))
}
