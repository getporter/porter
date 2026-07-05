//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
	"github.com/uwu-tools/magex/shx"
)

const fileSourcesBundleDir = "tests/integration/testdata/bundles/bundle-with-file-sources"

func TestFileSources_DownloadsSucceed(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	// Serve a file locally so the build does not require internet access.
	const fileContent = "hello from file server"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fileContent)
	}))
	defer srv.Close()

	fileURL := srv.URL + "/file.txt"

	// Write the bundle manifest directly with the test server URL baked in.
	porterYAML := fmt.Sprintf(`schemaVersion: 1.3.0
name: bundle-with-file-sources
version: 0.1.0
registry: localhost:5000

mixins:
  - exec

files:
  - url: %s
    destination: downloaded-file.txt

install:
  - exec:
      description: "Install"
      command: echo
      arguments:
        - installed

upgrade:
  - exec:
      description: "Upgrade"
      command: echo
      arguments:
        - upgraded

uninstall:
  - exec:
      description: "Uninstall"
      command: echo
      arguments:
        - uninstalled
`, fileURL)
	require.NoError(t, os.WriteFile(filepath.Join(test.TestDir, "porter.yaml"), []byte(porterYAML), 0644))
	test.Chdir(test.TestDir)

	_, output, err := test.RunPorterWith(func(cmd *shx.PreparedCommand) {
		cmd.Args("build", "--no-lint").
			Env("PORTER_EXPERIMENTAL=" + experimental.FileSources).
			Env("PORTER_ALLOW_FILE_DOWNLOADS=true")
	})
	require.NoError(t, err, "expected build to succeed with allow-file-downloads set")
	tests.RequireOutputContains(t, output, "Downloading "+fileURL)
}

func TestFileSources_BlockedWithoutAllowFlag(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	require.NoError(t, shx.Copy(filepath.Join(test.RepoRoot, fileSourcesBundleDir, "porter.yaml"), test.TestDir))
	test.Chdir(test.TestDir)

	_, _, err = test.RunPorterWith(func(cmd *shx.PreparedCommand) {
		cmd.Args("build", "--no-lint").
			Env("PORTER_EXPERIMENTAL=" + experimental.FileSources)
		// PORTER_ALLOW_FILE_DOWNLOADS intentionally not set
	})
	require.Error(t, err, "expected build to fail when allow-file-downloads is not set")
	require.ErrorContains(t, err, "--allow-file-downloads")
}

func TestFileSources_BlockedWithoutExperimentalFlag(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	require.NoError(t, shx.Copy(filepath.Join(test.RepoRoot, fileSourcesBundleDir, "porter.yaml"), test.TestDir))
	test.Chdir(test.TestDir)

	_, _, err = test.RunPorterWith(func(cmd *shx.PreparedCommand) {
		cmd.Args("build", "--no-lint")
		// PORTER_EXPERIMENTAL intentionally not set — schemaVersion 1.3.0 is unsupported without the flag
	})
	require.Error(t, err, "expected build to fail without file-sources experimental flag")
	require.ErrorContains(t, err, "invalid schema version")
}

func TestFileSources_WrongSchemaVersion(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	require.NoError(t, shx.Copy(filepath.Join(test.RepoRoot, fileSourcesBundleDir, "porter.yaml"), test.TestDir))
	test.Chdir(test.TestDir)

	// Downgrade the schema version so it is below the required 1.3.0.
	test.EditYaml("porter.yaml", func(yq *yaml.Editor) error {
		return yq.SetValue("schemaVersion", "1.2.0")
	})

	_, _, err = test.RunPorterWith(func(cmd *shx.PreparedCommand) {
		cmd.Args("build", "--no-lint").
			Env("PORTER_EXPERIMENTAL=" + experimental.FileSources)
	})
	require.Error(t, err, "expected build to fail with schema version below 1.3.0")
	require.ErrorContains(t, err, "schemaVersion >= 1.3.0")
}
