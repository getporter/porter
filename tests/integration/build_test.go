//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

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

	// Use a unique version for appversion so that docker doesn't cache the result and not print the value used in the Dockerfile
	appversion := fmt.Sprintf("app.versio=%d", time.Now().Unix())

	// build the bundle
	_, output := test.RequirePorter("build", "--custom", "customKey1=editedCustomValue1", "--custom", appversion, "--no-lint", "--name=porter-test-build")

	// Validate that the custom value defined in porter.yaml was injected into the build with --build-arg
	tests.RequireOutputContains(t, output, "CUSTOM_APP_VERSION=1.2.3")

	// Validate that the bundle metadata contains the custom key specified by the user with --custom
	bun, err := cnab.LoadBundle(test.TestContext.Context, build.LOCAL_BUNDLE)
	require.NoError(t, err)
	require.Equal(t, bun.Custom["customKey1"], "editedCustomValue1")

}

func TestRebuild(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Create a bundle
	test.Chdir(test.TestDir)
	test.RequirePorter("create")

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
