package porter

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/require"
)

func TestCreateInWorkingDirectory(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	err := p.Create()
	require.NoError(t, err)

	// Verify that files are present in the root directory
	configFileStats, err := p.FileSystem.Stat("porter.yaml")
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, "porter.yaml", pkg.FileModeWritable, configFileStats.Mode())

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

// tests to ensure behavior similarity with helm create
func TestCreateWithBundleName(t *testing.T) {
	bundleName := "mybundle"

	p := NewTestPorter(t)
	err := p.CreateInDir(bundleName)
	require.NoError(t, err)

	// Verify that files are present in the "mybundle" directory
	configFileStats, err := p.FileSystem.Stat(filepath.Join(bundleName, "porter.yaml"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(bundleName, "porter.yaml"), pkg.FileModeWritable, configFileStats.Mode())

	helperFileStats, err := p.FileSystem.Stat(filepath.Join(bundleName, "helpers.sh"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(bundleName, "helpers.sh"), pkg.FileModeExecutable, helperFileStats.Mode())

	dockerfileStats, err := p.FileSystem.Stat(filepath.Join(bundleName, "template.Dockerfile"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(bundleName, "template.Dockerfile"), pkg.FileModeWritable, dockerfileStats.Mode())

	readmeStats, err := p.FileSystem.Stat(filepath.Join(bundleName, "README.md"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(bundleName, "README.md"), pkg.FileModeWritable, readmeStats.Mode())

	gitignoreStats, err := p.FileSystem.Stat(filepath.Join(bundleName, ".gitignore"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(bundleName, ".gitignore"), pkg.FileModeWritable, gitignoreStats.Mode())

	dockerignoreStats, err := p.FileSystem.Stat(filepath.Join(bundleName, ".dockerignore"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(bundleName, ".dockerignore"), pkg.FileModeWritable, dockerignoreStats.Mode())

	// verify "name" inside porter.yaml is set to "mybundle"
	porterYaml := &manifest.Manifest{}
	data, err := p.FileSystem.ReadFile(filepath.Join(bundleName, "porter.yaml"))
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(data, &porterYaml))
	require.True(t, porterYaml.Name == bundleName)
}

// make sure bundlename is not the entire file structure, just the "base"
func TestCreateNestedBundleName(t *testing.T) {
	dir := "foo/bar/bar"
	bundleName := "mybundle"

	p := NewTestPorter(t)
	err := p.CreateInDir(filepath.Join(dir, bundleName))
	require.NoError(t, err)

	// Verify that files are present in the "mybundle" directory
	configFileStats, err := p.FileSystem.Stat(filepath.Join(dir, bundleName, "porter.yaml"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(dir, bundleName, "porter.yaml"), pkg.FileModeWritable, configFileStats.Mode())

	helperFileStats, err := p.FileSystem.Stat(filepath.Join(dir, bundleName, "helpers.sh"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(dir, bundleName, "helpers.sh"), pkg.FileModeExecutable, helperFileStats.Mode())

	dockerfileStats, err := p.FileSystem.Stat(filepath.Join(dir, bundleName, "template.Dockerfile"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(dir, bundleName, "template.Dockerfile"), pkg.FileModeWritable, dockerfileStats.Mode())

	readmeStats, err := p.FileSystem.Stat(filepath.Join(dir, bundleName, "README.md"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(dir, bundleName, "README.md"), pkg.FileModeWritable, readmeStats.Mode())

	gitignoreStats, err := p.FileSystem.Stat(filepath.Join(dir, bundleName, ".gitignore"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(dir, bundleName, ".gitignore"), pkg.FileModeWritable, gitignoreStats.Mode())

	dockerignoreStats, err := p.FileSystem.Stat(filepath.Join(dir, bundleName, ".dockerignore"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join(dir, bundleName, ".dockerignore"), pkg.FileModeWritable, dockerignoreStats.Mode())

	// verify "name" inside porter.yaml is set to "mybundle"
	porterYaml := &manifest.Manifest{}
	data, err := p.FileSystem.ReadFile(filepath.Join(dir, bundleName, "porter.yaml"))
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(data, &porterYaml))
	require.True(t, porterYaml.Name == bundleName)
}
