//go:build smoke

package smoke

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests/testdata"
	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/cnabio/cnab-go/packager"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/stretchr/testify/require"
)

// Validate that we can move a bundle into an aigraped environment
// and that it works without referencing the old environment/images.
// This also validates a lot of our insecure registry configuration.
func TestAirgappedEnvironment(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Start a temporary insecure test registry that uses self-signed certificates
	reg1 := test.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})

	// Publish referenced image to the insecure registry
	// This helps test that we can publish a bundle that references images from multiple registries
	// and that we properly apply skipTLS to those registries (and not just the registry to which we are pushing)
	referencedImg := "carolynvs/whalesayd@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f"
	localRegRepo := fmt.Sprintf("localhost:%s/whalesayd", reg1.Port)
	localRegRef := localRegRepo + ":latest"
	require.NoError(t, shx.RunE("docker", "pull", referencedImg))
	require.NoError(t, shx.RunE("docker", "tag", referencedImg, localRegRef))
	output, err := shx.OutputE("docker", "push", localRegRef)
	require.NoError(t, err)
	digest := getDigestFromDockerOutput(test.T, output)
	localRefWithDigest := fmt.Sprintf("%s@%s", localRegRepo, digest)

	// Start a second insecure test registry that uses self-signed certificates
	reg2 := test.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})

	// Edit the bundle so that it's referencing the image on the temporary registry
	// make sure the referenced image is not in local image cache
	shx.RunV("docker", "rmi", localRegRef)
	originTestBun := filepath.Join(test.RepoRoot, fmt.Sprintf("tests/testdata/%s/porter.yaml", testdata.MyBunsWithImgReference))
	testBun := filepath.Join(test.TestDir, "mybuns-img-reference.yaml")
	mgx.Must(shx.Copy(originTestBun, testBun))
	test.EditYaml(testBun, func(yq *yaml.Editor) error {
		return yq.SetValue("images.whalesayd.repository", fmt.Sprintf("localhost:%s/whalesayd", reg1.Port))
	})

	// Publish a test bundle that references the image from the temp registry, and push to another insecure registry
	test.RequirePorter("publish", "--file", "mybuns-img-reference.yaml", "--dir", test.TestDir, "--registry", reg2.String(), "--insecure-registry")
	reg1.Close()

	// Archive the bundle, it should not attempt to hit the first registry
	archiveFilePath := filepath.Join(test.TestDir, "archive-test.tgz")
	origRef := fmt.Sprintf("localhost:%s/%s:%s", reg2.Port, testdata.MyBunsWithImgReference, "v0.1.0")
	test.RequirePorter("archive", archiveFilePath, "--reference", origRef, "--insecure-registry")
	relocMap := getRelocationMap(test, archiveFilePath)
	require.Equal(test.T, fmt.Sprintf("localhost:%s/mybun-with-img-reference@sha256:499f71eec2e3bd78f26c268bbf5b2a65f73b96216fac4a89b86b5ebf115527b6", reg2.Port), relocMap[localRefWithDigest], "expected the relocation entry for the image to be the new published location")

	// Publish from the archived bundle to a new repository on the second registry
	newRef := fmt.Sprintf("localhost:%s/%s-second:%s", reg2.Port, testdata.MyBunsWithImgReference, "v0.2.0")
	test.RequirePorter("publish", "--archive", archiveFilePath, "-r", newRef, "--insecure-registry")
	archiveFilePath2 := filepath.Join(test.TestDir, "archive-test2.tgz")

	// Archive from the new location on the second registry
	test.RequirePorter("archive", archiveFilePath2, "--reference", newRef, "--insecure-registry")
	relocMap2 := getRelocationMap(test, archiveFilePath2)
	require.Equal(test.T, fmt.Sprintf("localhost:%s/mybun-with-img-reference-second@sha256:499f71eec2e3bd78f26c268bbf5b2a65f73b96216fac4a89b86b5ebf115527b6", reg2.Port), relocMap2[localRefWithDigest], "expected the relocation entry for the image to be the new published location")
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
