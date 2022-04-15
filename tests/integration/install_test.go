//go:build integration
// +build integration

package integration

import (
	"context"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstall_relativePathPorterHome(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest() // This creates a temp porter home directory
	p.Debug = false

	// Crux for this test: change Porter's home dir to a relative path
	homeDir, err := p.Config.GetHomeDir()
	require.NoError(t, err)
	relDir, err := filepath.Rel(p.Getwd(), homeDir)
	require.NoError(t, err)
	p.SetHomeDir(relDir)

	// Bring in a porter manifest that has an install action defined
	p.AddTestBundleDir("testdata/bundles/bundle-with-custom-action", true)

	installOpts := porter.NewInstallOptions()
	err = installOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)

	// Install the bundle, assert no error occurs due to Porter home as relative path
	err = p.InstallBundle(context.Background(), installOpts)
	require.NoError(t, err)
}

func TestInstall_fileParam(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	p.Debug = false

	bundleName := p.AddTestBundleDir("testdata/bundles/bundle-with-file-params", false)

	installOpts := porter.NewInstallOptions()
	installOpts.Params = []string{"myfile=./myfile"}
	installOpts.ParameterSets = []string{"myparam"}
	testParamSets := parameters.NewParameterSet("", "myparam", secrets.Strategy{
		Name: "myotherfile",
		Source: secrets.Source{
			Key:   host.SourcePath,
			Value: "./myotherfile",
		},
	})

	p.TestParameters.InsertParameterSet(testParamSets)

	err := installOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(context.Background(), installOpts)
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	require.Contains(t, output, "Hello World!", "expected action output to contain provided file contents")

	outputs, err := p.Claims.GetLastOutputs("", bundleName)
	require.NoError(t, err, "GetLastOutput failed")
	myfile, ok := outputs.GetByName("myfile")
	require.True(t, ok, "expected myfile output to be persisted")
	assert.Equal(t, "Hello World!", string(myfile.Value), "expected output to match the decoded file contents")
	myotherfile, ok := outputs.GetByName("myotherfile")
	require.True(t, ok, "expected myotherfile output to be persisted")
	assert.Equal(t, "Hello Other World!", string(myotherfile.Value), "expected output 'myotherfile' to match the decoded file contents")
}

func TestInstall_withDockerignore(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	p.Debug = false

	p.AddTestBundleDir("testdata/bundles/outputs-example", true)

	// Create .dockerignore file which ignores the helpers script
	err := p.FileSystem.WriteFile(".dockerignore", []byte("helpers.sh"), pkg.FileModeWritable)
	require.NoError(t, err)

	opts := porter.NewInstallOptions()
	err = opts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)

	// Verify Porter uses the .dockerignore file (no helpers script added to installer image)
	err = p.InstallBundle(context.Background(), opts)
	// The following line would be seen from the daemon, but is printed directly to stdout:
	// Error: couldn't run command ./helpers.sh dump-config: fork/exec ./helpers.sh: no such file or directory
	// We should check this once https://github.com/cnabio/cnab-go/issues/78 is closed
	require.EqualError(t, err, "1 error occurred:\n\t* container exit code: 1, message: <nil>\n\n")
}

func TestInstall_stringParam(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	p.Debug = false

	p.AddTestBundleDir("testdata/bundles/bundle-with-string-params", false)

	installOpts := porter.NewInstallOptions()
	installOpts.Params = []string{"name=Demo Time"}

	err := installOpts.Validate([]string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(context.Background(), installOpts)
	require.NoError(t, err)

	output := p.TestConfig.TestContext.GetOutput()
	require.Contains(t, output, "Hello, Demo Time", "expected action output to contain provided file contents")
}
