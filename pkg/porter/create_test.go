package porter

import (
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/require"
)

func TestCreateInRootDirectory(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	err := p.Create("")
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

func TestCreateInDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := os.TempDir()
	bundleDir := filepath.Join(tempDir, "mybundle")

	p := NewTestPorter(t)
	err := p.Create(bundleDir)
	require.NoError(t, err)

	// Verify that files are present in the "mybundle" directory
	configFileStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, "porter.yaml"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("mybundle", "porter.yaml"), pkg.FileModeWritable, configFileStats.Mode())

	helperFileStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, "helpers.sh"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("mybundle", "helpers.sh"), pkg.FileModeExecutable, helperFileStats.Mode())

	dockerfileStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, "template.Dockerfile"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("mybundle", "template.Dockerfile"), pkg.FileModeWritable, dockerfileStats.Mode())

	readmeStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, "README.md"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("mybundle", "README.md"), pkg.FileModeWritable, readmeStats.Mode())

	gitignoreStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, ".gitignore"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("mybundle", ".gitignore"), pkg.FileModeWritable, gitignoreStats.Mode())

	dockerignoreStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, ".dockerignore"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("mybundle", ".dockerignore"), pkg.FileModeWritable, dockerignoreStats.Mode())
}

func TestCreateInChildDirectoryWithoutExistingParentDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := os.TempDir()
	parentDir := filepath.Join(tempDir, "parentbundle")
	err := os.Mkdir(parentDir, os.ModePerm)
	require.NoError(t, err)

	// Define the child directory within the existing parent directory
	bundleDir := filepath.Join(parentDir, "childbundle")

	p := NewTestPorter(t)
	err = p.Create(bundleDir)

	// Verify that the expected error is returned
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create directory for bundle")

	// Verify that no files are created in the subdirectory
	_, err = p.FileSystem.Stat(filepath.Join(bundleDir, "porter.yaml"))
	require.True(t, os.IsNotExist(err))

	_, err = p.FileSystem.Stat(filepath.Join(bundleDir, "helpers.sh"))
	require.True(t, os.IsNotExist(err))

	_, err = p.FileSystem.Stat(filepath.Join(bundleDir, "template.Dockerfile"))
	require.True(t, os.IsNotExist(err))

	_, err = p.FileSystem.Stat(filepath.Join(bundleDir, "README.md"))
	require.True(t, os.IsNotExist(err))

	_, err = p.FileSystem.Stat(filepath.Join(bundleDir, ".gitignore"))
	require.True(t, os.IsNotExist(err))

	_, err = p.FileSystem.Stat(filepath.Join(bundleDir, ".dockerignore"))
	require.True(t, os.IsNotExist(err))
}

func TestCreateInChildDirectoryWithExistingParentDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := os.TempDir()
	parentDir := filepath.Join(tempDir, "parentbundle")
	err := os.Mkdir(parentDir, os.ModePerm)
	require.NoError(t, err)

	// Define the child directory within the existing parent directory
	bundleDir := filepath.Join(parentDir, "childbundle")

	p := NewTestPorter(t)
	err = p.Create(bundleDir)
	require.NoError(t, err)

	// Verify that files are present in the child directory
	configFileStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, "porter.yaml"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("parentbundle/childbundle", "porter.yaml"), pkg.FileModeWritable, configFileStats.Mode())

	helperFileStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, "helpers.sh"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("parentbundle/childbundle", "helpers.sh"), pkg.FileModeExecutable, helperFileStats.Mode())

	dockerfileStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, "template.Dockerfile"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("parentbundle/childbundle", "template.Dockerfile"), pkg.FileModeWritable, dockerfileStats.Mode())

	readmeStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, "README.md"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("parentbundle/childbundle", "README.md"), pkg.FileModeWritable, readmeStats.Mode())

	gitignoreStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, ".gitignore"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("parentbundle/childbundle", ".gitignore"), pkg.FileModeWritable, gitignoreStats.Mode())

	dockerignoreStats, err := p.FileSystem.Stat(filepath.Join(bundleDir, ".dockerignore"))
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, filepath.Join("parentbundle/childbundle", ".dockerignore"), pkg.FileModeWritable, dockerignoreStats.Mode())
}
