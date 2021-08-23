// +build smoke

package smoke

import (
	"encoding/json"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg/claims"
	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/require"
)

const (
	myBunsRef = "localhost:5000/mybuns:v0.1.2"
	myDbRef   = "localhost:5000/mydb:v0.1.0"
)

func (t Test) PrepareTestBundle() {
	// This variable isn't set on windows and the mybuns bundle relies on it
	os.Setenv("USER", "porterci")

	// Check if another test has already set up the test bundle
	err := shx.RunE("docker", "pull", "localhost:5000/mybuns-installer:v0.1.2")
	if err == nil {
		return
	}

	// Build and publish an interesting test bundle and its dependency
	t.MakeTestBundle("mydb", myDbRef)
	t.MakeTestBundle("mybuns", myBunsRef)
}

func (t Test) MakeTestBundle(name string, ref string) {
	err := shx.Copy(filepath.Join(t.RepoRoot, "tests/testdata", name), t.TestDir, shx.CopyRecursive)
	require.NoError(t.T, err)

	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)
	os.Chdir(filepath.Join(t.TestDir, name))

	t.RequirePorter("build")
	t.RequirePorter("publish", "--reference", ref)
}

func (t Test) ShowInstallation(namespace string, name string) (claims.Installation, error) {
	output, err := t.RunPorter("show", name, "--namespace", namespace, "--output=json")
	if err != nil {
		return claims.Installation{}, err
	}

	var installation claims.Installation
	require.NoError(t.T, json.Unmarshal([]byte(output), &installation))
	return installation, nil
}

func (t Test) RequireInstallationExists(namespace string, name string) claims.Installation {
	installation, err := t.ShowInstallation(namespace, name)
	require.NoError(t.T, err)
	require.Equal(t.T, name, installation.Name, "incorrect installation name")
	require.Equal(t.T, namespace, installation.Namespace, "incorrect installation namespace")
	return installation
}

func (t Test) RequireInstallationNotFound(namespace string, name string) {
	_, err := t.ShowInstallation(namespace, name)
	t.RequireNotFoundReturned(err)
}

func (t Test) RequireNotFoundReturned(err error) {
	require.Error(t.T, err)
	require.Contains(t.T, err.Error(), "not found")
}

func (t Test) ListInstallations(allNamespaces bool, namespace string, name string, labels []string) ([]claims.Installation, error) {
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

	output, err := t.RunPorter(args...)
	if err != nil {
		return nil, err
	}

	var installations []claims.Installation
	require.NoError(t.T, json.Unmarshal([]byte(output), &installations))
	return installations, nil
}

func (t Test) RequireInstallationInList(namespace, name string, list []claims.Installation) claims.Installation {
	for _, i := range list {
		if i.Namespace == namespace && i.Name == name {
			return i
		}
	}

	t.T.Fatalf("expected %s/%s to be in the list of installations", namespace, name)
	return claims.Installation{}
}
