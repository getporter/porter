package tester

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/testdata"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

// PrepareTestBundle ensures that the mybuns test bundle is ready to use.
func (t Tester) PrepareTestBundle() {
	// Build and publish an interesting test bundle and its dependency
	t.MakeTestBundle(testdata.MyDb, testdata.MyDbRef)
	t.MakeTestBundle(testdata.MyBuns, testdata.MyBunsRef)

	t.ApplyTestBundlePrerequisites()
}

// ApplyTestBundlePrerequisites ensures that anything required by the test bundle, mybuns, is ready to use.
func (t Tester) ApplyTestBundlePrerequisites() {
	// These are environment variables referenced by the mybuns credential set
	os.Setenv("USER", "porterci")
	os.Setenv("ALT_USER", "porterci2")

	t.RequirePorter("parameters", "apply", filepath.Join(t.RepoRoot, "tests/testdata/params/mybuns.yaml"), "--namespace=")
	t.RequirePorter("credentials", "apply", filepath.Join(t.RepoRoot, "tests/testdata/creds/mybuns.yaml"), "--namespace=")
}

func (t Tester) MakeTestBundle(name string, ref string) {
	// Skip if we've already pushed it for another test
	if _, _, err := t.RunPorter("explain", ref); err == nil {
		return
	}

	pwd, _ := os.Getwd()
	defer t.Chdir(pwd)
	t.Chdir(filepath.Join(t.RepoRoot, "tests/testdata/", name))

	// TODO(carolynvs): porter publish detection of needing a build should do this
	output, err := shx.OutputS("docker", "inspect", strings.Replace(ref, name, name+"-installer", 1))
	if output == "[]" || err != nil {
		t.RequirePorter("build")
	}

	// Rely on the auto build functionality to avoid long slow rebuilds when nothing has changed
	t.RequirePorter("publish", "--reference", ref)
}

func (t Tester) ShowInstallation(namespace string, name string) (porter.DisplayInstallation, error) {
	stdout, _, err := t.RunPorter("show", name, "--namespace", namespace, "--output=json")
	if err != nil {
		return porter.DisplayInstallation{}, err
	}

	var di porter.DisplayInstallation

	require.NoError(t.T, json.Unmarshal([]byte(stdout), &di))

	return di, nil
}

func (t Tester) RequireInstallationExists(namespace string, name string) porter.DisplayInstallation {
	di, err := t.ShowInstallation(namespace, name)
	require.NoError(t.T, err)
	require.Equal(t.T, name, di.Name, "incorrect installation name")
	require.Equal(t.T, namespace, di.Namespace, "incorrect installation namespace")

	return di
}

func (t Tester) RequireInstallationNotFound(namespace string, name string) {
	_, err := t.ShowInstallation(namespace, name)
	t.RequireNotFoundReturned(err)
}

func (t Tester) RequireNotFoundReturned(err error) {
	require.Error(t.T, err)
	require.Contains(t.T, err.Error(), "not found")
}

func (t Tester) ListInstallations(allNamespaces bool, namespace string, name string, labels []string) ([]porter.DisplayInstallation, error) {
	args := []string{
		"list",
		"--output=json",
		"--name", name,
	}
	if allNamespaces {
		args = append(args, "--all-namespaces")
	} else {
		args = append(args, "--namespace", namespace)
	}
	for _, l := range labels {
		args = append(args, "--label", l)
	}

	stdout, _, err := t.RunPorter(args...)
	if err != nil {
		return nil, err
	}

	var installations []porter.DisplayInstallation
	require.NoError(t.T, json.Unmarshal([]byte(stdout), &installations))
	return installations, nil
}

func (t Tester) RequireInstallationInList(namespace, name string, list []storage.Installation) storage.Installation {
	for _, i := range list {
		if i.Namespace == namespace && i.Name == name {
			return i
		}
	}

	t.T.Fatalf("expected %s/%s to be in the list of installations", namespace, name)
	return storage.Installation{}
}

// EditYaml applies a set of yq transformations to a file.
func (t Tester) EditYaml(path string, transformations ...func(yq *yaml.Editor) error) {
	t.T.Log("Editing", path)
	yq := yaml.NewEditor(t.TestContext.Context)

	require.NoError(t.T, yq.ReadFile(path))
	for _, transform := range transformations {
		require.NoError(t.T, transform(yq))
	}
	require.NoError(t.T, yq.WriteFile(path))
}

// RequireFileMode checks that all files in the specified path match the specifed
// file mode. Uses a glob pattern to match.
func (t *Tester) RequireFileMode(path string, mode os.FileMode) {
	if !tests.AssertDirectoryPermissionsEqual(t.T, path, mode) {
		t.T.FailNow()
	}
}
