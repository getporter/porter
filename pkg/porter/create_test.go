package porter

import (
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	err := p.Create()
	require.NoError(t, err)

	configFileStats, err := p.FileSystem.Stat("porter.yaml")
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, "porter.yaml", pkg.FileModeWritable, configFileStats.Mode())

	// Verify that helpers is present and executable
	helperFileStats, err := p.FileSystem.Stat("helpers.sh")
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, "helpers.sh", pkg.FileModeExecutable, helperFileStats.Mode())

	dockerfileStats, err := p.FileSystem.Stat("template.Dockerfile")
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, "template.Dockerfile", pkg.FileModeWritable, dockerfileStats.Mode())

	readmeStats, err := p.FileSystem.Stat("README.md")
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, "README.md", pkg.FileModeWritable, readmeStats.Mode())

	gitignoreStats, err := p.FileSystem.Stat(".gitignore")
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, ".gitignore", pkg.FileModeWritable, gitignoreStats.Mode())

	dockerignoreStats, err := p.FileSystem.Stat(".dockerignore")
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, ".dockerignore", pkg.FileModeWritable, dockerignoreStats.Mode())

}
