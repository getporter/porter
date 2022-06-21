//go:build smoke
// +build smoke

package smoke

import (
	"fmt"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests/testdata"
	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// Start up another docker registry to host the original bundle
// Publish a bundle to the temporary registry
// Archive the bundle from the temporary registry
// Stop the temporary registry
// Publish the bundle from the archive file
func TestArchive(t *testing.T) {
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
	localRegRef := fmt.Sprintf("localhost:%s/whalesayd:latest", tempRegistryPort)
	require.NoError(t, shx.RunE("docker", "pull", "carolynvs/whalesayd@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f"))
	require.NoError(t, shx.RunE("docker", "tag", "carolynvs/whalesayd@sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f", localRegRef))
	require.NoError(t, shx.RunE("docker", "push", localRegRef))

	// publish a test bundle that reference the image from the temp registry
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
}
