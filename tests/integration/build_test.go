//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/tester"
	"github.com/Masterminds/semver/v3"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	bunPath := filepath.Join(test.RepoRoot, "tests/testdata/mybuns/*")
	require.NoError(t, shx.Copy(bunPath, test.TestDir, shx.CopyRecursive))
	test.Chdir(test.TestDir)

	// build the bundle
	_, output := test.RequirePorter("build", "--custom", "customKey1=editedCustomValue1", "--no-lint", "--name=porter-test-build")

	// Validate that the custom value defined in porter.yaml was injected into the build as a build argument
	tests.RequireOutputContains(t, output, "CUSTOM_APP_VERSION=1.2.3")

	// Validate that the bundle metadata contains the custom key specified by the user with --custom
	bun, err := cnab.LoadBundle(test.TestContext.Context, build.LOCAL_BUNDLE)
	require.NoError(t, err)
	require.Equal(t, bun.Custom["customKey1"], "editedCustomValue1")

}

// This test uses build and the --no-lint flag, which is not a global flag defined on config.DataStore,
// to validate that when a flag value is set via the configuration file, environment variable or directly with the flag
// that the value binds properly to the variable.
// It is a regression test for our cobra+viper configuration setup and was created in response to this regression
// https://github.com/getporter/porter/issues/2735
func TestBuild_ConfigureNoLintThreeWays(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Use a bundle that will trigger a lint error, it can only be successfully built when --no-lint is set
	require.NoError(t, shx.Copy("testdata/bundles/bundle-with-lint-error/porter.yaml", test.TestDir))
	test.Chdir(test.TestDir)

	// Attempt to build the bundle, it should fail with a lint error
	_, _, err = test.RunPorter("build")
	require.ErrorContains(t, err, "lint errors were detected")

	// Build the bundle and disable the linter with --no-lint
	test.RequirePorter("build", "--no-lint")

	// Build the bundle and disable the linter with PORTER_NO_LINT
	_, _, err = test.RunPorterWith(func(cmd *shx.PreparedCommand) {
		cmd.Args("build").Env("PORTER_NO_LINT=true")
	})
	require.NoError(t, err, "expected the build to pass when PORTER_NO_LINT=true is specified")

	// Build the bundle and disable the linter with the configuration file
	disableAutoBuildCfg := []byte(`no-lint: true`)
	err = os.WriteFile(filepath.Join(test.PorterHomeDir, "config.yaml"), disableAutoBuildCfg, pkg.FileModeWritable)
	require.NoError(t, err, "Failed to write out the porter configuration file")
	test.RequirePorter("build")
}

func TestRebuild(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Create a bundle
	test.Chdir(test.TestDir)
	test.RequirePorter("create")

	// Edit the bundle to use more than one mixin
	// This helps us test that when we calculate the manifestDigest that it consistently sorts used mixins
	test.EditYaml("porter.yaml", func(yq *yaml.Editor) error {
		n, err := yq.GetNode("mixins")
		require.NoError(t, err, "could not get mixins node for porter.yaml")
		testMixin := *n.Content[0]
		testMixin.Value = "testmixin"
		n.Content = append(n.Content, &testMixin)
		return nil
	})

	// Use a unique name with it so that if other tests install a newly created hello
	// world bundle, they don't conflict
	installationName := t.Name()

	// Modify the porter.yaml to trigger a rebuild
	bumpBundle := func() {
		test.EditYaml("porter.yaml", func(yq *yaml.Editor) error {
			orig, err := yq.GetValue("version")
			require.NoError(t, err, "unable to read the bundle version from porter.yaml in order to bump it")

			v, err := semver.NewVersion(orig)
			require.NoErrorf(t, err, "error reading %s as a semver version", orig)

			return yq.SetValue("version", v.IncPatch().String())
		})
	}

	// Try to explain the bundle without building first, it should fail
	_, output, err := test.RunPorter("explain", "--autobuild-disabled")
	require.ErrorContains(t, err, "Attempted to use a bundle from source without building it first when --autobuild-disabled is set. Build the bundle and try again")
	require.NotContains(t, output, "Building bundle ===>")

	// Explain the bundle
	_, output = test.RequirePorter("explain")
	tests.RequireOutputContains(t, output, "Building bundle ===>", "expected a build before explain")

	// Explain the bundle a bunch more times, it should not rebuild again.
	// This is a regression test for a bug where the manifest would be considered out-of-date when nothing had changed
	// caused by us using a go map when comparing the mixins used in the bundle, which has inconsistent sort order...

	//todo: This test is flaky still and upsetting CI
	// for i := 0; i < 5; i++ {
	// 	_, output = test.RequirePorter("explain")
	// 	tests.RequireOutputContains(t, output, "Bundle is up-to-date!", "expected the previous build to be reused")
	// }

	bumpBundle()

	// Explain the bundle, with --autobuild-disabled. It should work since the bundle has been built
	explainJson, output := test.RequirePorter("explain", "--autobuild-disabled", "-o=json")
	tests.RequireOutputContains(t, output, "WARNING: The bundle is out-of-date. Skipping autobuild because --autobuild-disabled was specified")
	require.NotContains(t, output, "Building bundle ===>")
	var explainResult map[string]interface{}
	err = json.Unmarshal([]byte(explainJson), &explainResult)
	require.NoError(t, err, "could not marshal explain output as json")
	require.Equal(t, "0.1.0", explainResult["version"], "explain should show stale output because we used --autobuild-disabled")

	// Inspect the bundle
	_, output = test.RequirePorter("inspect")
	tests.RequireOutputContains(t, output, "Building bundle ===>", "expected a build before inspect")

	bumpBundle()

	// Generate credentials for the bundle
	_, output = test.RequirePorter("credentials", "generate", installationName)
	tests.RequireOutputContains(t, output, "Building bundle ===>", "expected a build before credentials generate")

	bumpBundle()

	// Generate parameters for the bundle
	_, output = test.RequirePorter("parameters", "generate", installationName)
	tests.RequireOutputContains(t, output, "Building bundle ===>", "expected a build before parameters generate")

	bumpBundle()

	// Install the bundle
	_, output = test.RequirePorter("install", installationName)
	tests.RequireOutputContains(t, output, "Building bundle ===>", "expected a build before install")

	bumpBundle()

	// Upgrade the bundle
	_, output = test.RequirePorter("upgrade", installationName)
	tests.RequireOutputContains(t, output, "Building bundle ===>", "expected a rebuild before upgrade")

	// Upgrade again, should not trigger a rebuild
	_, output = test.RequirePorter("upgrade", installationName)
	require.NotContains(t, output, "Building bundle ===>", "the second upgrade should not rebuild because the bundle wasn't changed")

	bumpBundle()

	// Uninstall the bundle
	_, output = test.RequirePorter("uninstall", installationName)
	tests.RequireOutputContains(t, output, "Building bundle ===>", "expected a rebuild before uninstall")
}
