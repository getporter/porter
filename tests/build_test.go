// +build integration

package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"get.porter.sh/porter/pkg/porter"
)

func TestBuild_withDockerignore(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(p.TestDir, "testdata/bundles/outputs-example"), ".")

	// Create .dockerignore file which ignores the Dockerfile
	err := p.FileSystem.WriteFile(".dockerignore", []byte("Dockerfile"), 0644)
	require.NoError(t, err)

	// Verify Porter uses the .dockerignore file
	opts := porter.BuildOptions{}
	err = p.Build(opts)
	require.EqualError(t, err, "unable to build CNAB invocation image: Error response from daemon: Cannot locate specified Dockerfile: Dockerfile")
}
