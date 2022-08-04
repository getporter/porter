//go:build integration

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
	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/testdata"
	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/cnabio/cnab-go/packager"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Validate that when we archive a bundle, everything is included inside
// We should be able to archive, move to a disconnected environment
// and publish again without accidentally referencing artifacts from the original environment.
func TestArchive_AirgappedEnvironment(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Start a temp registry
	tempRegistryId, err := shx.OutputE("docker", "run", "-d", "-P", "registry:2")
	require.NoError(t, err, "Could not start a temporary registry")
	stopTempRegistry := func() error {
		return shx.RunE("docker", "rm", "-vf", tempRegistryId)
	}
	defer stopTempRegistry()

	// Get the port that it is running on
	tempRegistryPort, err := shx.OutputE("docker", "inspect", tempRegistryId, "--format", `{{ (index (index .NetworkSettings.Ports "5000/tcp") 0).HostPort }}`)
	require.NoError(t, err, "Could not get the published port of the temporary registry")

	// Publish referenced bundle to one location
	referencedImg := "carolynvs/whalesayd@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f"
	localRegRepo := fmt.Sprintf("localhost:%s/whalesayd", tempRegistryPort)
	localRegRef := localRegRepo + ":latest"
	require.NoError(t, shx.RunE("docker", "pull", referencedImg))
	require.NoError(t, shx.RunE("docker", "tag", referencedImg, localRegRef))
	output, err := shx.OutputE("docker", "push", localRegRef)
	require.NoError(t, err)
	digest := getDigestFromDockerOutput(test.T, output)
	localRefWithDigest := fmt.Sprintf("%s@%s", localRegRepo, digest)

	// publish a test bundle that reference the image from the temp registry
	// make sure the referenced image is not in local image cache
	shx.RunV("docker", "rmi", localRegRef)
	originTestBun := filepath.Join(test.RepoRoot, fmt.Sprintf("tests/testdata/%s/porter.yaml", testdata.MyBunsWithImgReference))
	testBun := filepath.Join(test.TestDir, "mybuns-img-reference.yaml")
	mgx.Must(shx.Copy(originTestBun, testBun))
	test.EditYaml(testBun, func(yq *yaml.Editor) error {
		return yq.SetValue("images.whalesayd.repository", fmt.Sprintf("localhost:%s/whalesayd", tempRegistryPort))
	})
	test.RequirePorter("publish", "--file", "mybuns-img-reference.yaml", "--dir", test.TestDir)
	stopTempRegistry()

	archiveFilePath := filepath.Join(test.TestDir, "archive-test.tgz")
	test.RequirePorter("archive", archiveFilePath, "--reference", testdata.MyBunsWithImgReferenceRef)
	relocMap := getRelocationMap(test, archiveFilePath)
	require.Equal(test.T, relocMap[localRefWithDigest], "localhost:5000/mybun-with-img-reference@sha256:499f71eec2e3bd78f26c268bbf5b2a65f73b96216fac4a89b86b5ebf115527b6", relocMap)

	// publish from the archived bundle
	newRef := fmt.Sprintf("localhost:5000/%s-%s:%s", testdata.MyBunsWithImgReference, "second", "v0.2.0")
	test.RequirePorter("publish", "--archive", archiveFilePath, "-r", newRef)
	archiveFilePath2 := filepath.Join(test.TestDir, "archive-test2.tgz")
	test.RequirePorter("archive", archiveFilePath2, "--reference", newRef)
	relocMap2 := getRelocationMap(test, archiveFilePath2)
	require.Equal(test.T, relocMap2[localRefWithDigest], "localhost:5000/mybun-with-img-reference-second@sha256:499f71eec2e3bd78f26c268bbf5b2a65f73b96216fac4a89b86b5ebf115527b6", relocMap)
}

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
	err = publishFromArchiveOpts.Validate(p.Context)
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

func getRelocationMap(test tester.Tester, archiveFilePath string) relocation.ImageRelocationMap {
	l := loader.NewLoader()
	imp := packager.NewImporter(archiveFilePath, test.TestDir, l)
	err := imp.Import()
	require.NoError(test.T, err, "opening archive failed")

	_, err = test.TestContext.FileSystem.Stat(filepath.Join(test.TestDir, strings.TrimSuffix(filepath.Base(archiveFilePath), ".tgz"), "bundle.json"))
	require.NoError(test.T, err)
	relocMapBytes, err := test.TestContext.FileSystem.ReadFile(filepath.Join(test.TestDir, strings.TrimSuffix(filepath.Base(archiveFilePath), ".tgz"), "relocation-mapping.json"))
	require.NoError(test.T, err)

	// make sure the relocation map contains the expected image
	relocMap := relocation.ImageRelocationMap{}
	require.NoError(test.T, json.Unmarshal(relocMapBytes, &relocMap))
	return relocMap
}

func getDigestFromDockerOutput(t *testing.T, output string) string {
	_, after, found := strings.Cut(output, "digest: ")
	require.True(t, found)
	results := strings.Split(after, " ")
	require.Greater(t, len(results), 1)
	require.Contains(t, results[0], "sha256")

	return results[0]
}
