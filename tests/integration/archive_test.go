//go:build integration
// +build integration

package integration

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/cnabio/cnab-go/packager"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchive(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()
	p.Debug = false

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
	containsRequiredMetadata(p, archiveFile1)

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
	containsRequiredMetadata(p, archiveFile2)

	// Publish bundle from archive, with new reference
	publishFromArchiveOpts := porter.PublishOptions{
		ArchiveFile: archiveFile1,
		BundlePullOptions: porter.BundlePullOptions{
			Reference: fmt.Sprintf("localhost:5000/archived-whalegap:v0.2.0"),
		},
	}
	err = publishFromArchiveOpts.Validate(p.Context)
	require.NoError(p.T(), err, "validation of publish opts for bundle failed")

	err = p.Publish(ctx, publishFromArchiveOpts)
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

func containsRequiredMetadata(p *porter.TestPorter, path string) {
	tmpDir, err := p.FileSystem.TempDir("", "porter-integration-tests")
	require.NoError(p.T(), err)
	defer p.FileSystem.RemoveAll(tmpDir)

	source := p.FileSystem.Abs(path)
	l := loader.NewLoader()
	imp := packager.NewImporter(source, tmpDir, l)
	err = imp.Import()
	require.NoError(p.T(), err, "opening archive failed")

	relocMapBytes, err := p.FileSystem.ReadFile(filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"), "relocation-mapping.json"))
	require.NoError(p.T(), err)

	// make sure the relocation map contains the expected image
	relocMap := relocation.ImageRelocationMap{}
	require.NoError(p.T(), json.Unmarshal(relocMapBytes, &relocMap))
	require.Equal(p.T(), relocMap["carolynvs/whalesayd@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f"], "ghcr.io/getporter/examples/whalegap@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f", relocMap)
	require.Equal(p.T(), relocMap["ghcr.io/getporter/examples/whalegap:2bed6d7bf087c159835ddfac5838fd34@sha256:5ada057d9b24c443d9fb0240b49c13a5a36a11d5f4dda8adaaa2ec74e39c0d24"], "ghcr.io/getporter/examples/whalegap@sha256:5ada057d9b24c443d9fb0240b49c13a5a36a11d5f4dda8adaaa2ec74e39c0d24", relocMap)

	_, err = p.FileSystem.Stat(filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"), "bundle.json"))
	require.NoError(p.T(), err)
}
