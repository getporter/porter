// +build integration

package integration

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchive(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	// Use a fixed bundle to work with so that we can rely on the registry and layer digests
	const reference = "getporter/mysql:v0.1.4"

	// Archive bundle
	archive1Opts := porter.ArchiveOptions{}
	archive1Opts.Reference = reference
	archiveFile1 := "mybuns1.tgz"
	err := archive1Opts.Validate([]string{archiveFile1}, p.Porter)
	require.NoError(p.T(), err, "validation of archive opts for bundle failed")

	err = p.Archive(archive1Opts)
	require.NoError(p.T(), err, "archival of bundle failed")

	info, err := p.FileSystem.Stat(archiveFile1)
	require.NoError(p.T(), err)
	require.Equal(p.T(), os.FileMode(0644), info.Mode())

	hash1 := getHash(p, archiveFile1)

	// Check to be sure the shasum is stable after archiving a second time
	archive2Opts := porter.ArchiveOptions{}
	archive2Opts.Reference = reference
	archiveFile2 := "mybuns2.tgz"
	err = archive2Opts.Validate([]string{archiveFile2}, p.Porter)
	require.NoError(p.T(), err, "validation of archive opts for bundle failed")

	err = archive1Opts.Validate([]string{archiveFile2}, p.Porter)
	require.NoError(t, err, "Second validate failed")

	err = p.Archive(archive2Opts)
	require.NoError(t, err, "Second archive failed")

	assert.Equal(p.T(), hash1, getHash(p, archiveFile2), "shasum of archive did not stay the same on the second call to archive")

	// Publish bundle from archive, with new reference
	publishFromArchiveOpts := porter.PublishOptions{
		ArchiveFile: archiveFile1,
		BundlePullOptions: porter.BundlePullOptions{
			Reference: fmt.Sprintf("localhost:5000/archived-mysql:v0.1.4"),
		},
	}
	err = publishFromArchiveOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for bundle failed")

	err = p.Publish(publishFromArchiveOpts)
	require.NoError(p.T(), err, "publish of bundle from archive failed")
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
