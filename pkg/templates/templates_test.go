package templates

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplates_GetManifest(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := NewTemplates(c.Config)

	gotTmpl, err := tmpl.GetManifest()
	require.NoError(t, err)

	wantTmpl, _ := os.ReadFile("./templates/create/porter.yaml")
	assert.Equal(t, string(wantTmpl), string(gotTmpl))
}

func TestTemplates_GetRunScript(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := NewTemplates(c.Config)

	gotTmpl, err := tmpl.GetRunScript()
	require.NoError(t, err)

	wantTmpl, _ := os.ReadFile("./templates/build/cnab/app/run")
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestTemplates_GetDockerfile(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := NewTemplates(c.Config)

	gotTmpl, err := tmpl.GetDockerfile()
	require.NoError(t, err)

	strTmpl := string(gotTmpl)
	require.Contains(t, strTmpl, "--platform=linux/amd64", "missing default platform flag")
	test.CompareGoldenFile(t, "./templates/build/buildkit.Dockerfile", strTmpl)
}

func TestTemplates_GetDockerfileTemplate(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := NewTemplates(c.Config)

	gotTmpl, err := tmpl.GetDockerfileTemplate()
	require.NoError(t, err)

	strTmpl := string(gotTmpl)
	require.Contains(t, strTmpl, "--platform=linux/amd64", "missing default platform flag")
	test.CompareGoldenFile(t, "./templates/create/template.buildkit.Dockerfile", strTmpl)
}

func TestTemplates_GetCredentialSetJSON(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := NewTemplates(c.Config)

	gotTmpl, err := tmpl.GetCredentialSetJSON()
	require.NoError(t, err)

	wantTmpl, err := os.ReadFile("./templates/credentials/create/credential-set.json")
	require.NoError(t, err)
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestTemplates_GetCredentialSetYAML(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := NewTemplates(c.Config)

	gotTmpl, err := tmpl.GetCredentialSetYAML()
	require.NoError(t, err)

	wantTmpl, err := os.ReadFile("./templates/credentials/create/credential-set.yaml")
	require.NoError(t, err)
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestTemplates_GetParameterSetJSON(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := NewTemplates(c.Config)

	gotTmpl, err := tmpl.GetParameterSetJSON()
	require.NoError(t, err)

	wantTmpl, err := os.ReadFile("./templates/parameters/create/parameter-set.json")
	require.NoError(t, err)
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestTemplates_GetParameterSetYAML(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := NewTemplates(c.Config)

	gotTmpl, err := tmpl.GetParameterSetYAML()
	require.NoError(t, err)

	wantTmpl, err := os.ReadFile("./templates/parameters/create/parameter-set.yaml")
	require.NoError(t, err)
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestTemplates_GetSchema(t *testing.T) {
	t.Run("stable schema", func(t *testing.T) {
		c := config.NewTestConfig(t)
		tmpl := NewTemplates(c.Config)

		gotTmpl, err := tmpl.GetSchema()
		require.NoError(t, err)

		wantTmpl, err := os.ReadFile("./templates/schema.json")
		require.NoError(t, err)
		require.Equal(t, string(wantTmpl), string(gotTmpl))
	})

	t.Run("experimental schema", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetExperimentalFlags(experimental.FlagDependenciesV2)
		tmpl := NewTemplates(c.Config)

		gotTmpl, err := tmpl.GetSchema()
		require.NoError(t, err)

		wantTmpl, err := os.ReadFile("./templates/v1.1.0.schema.json")
		require.NoError(t, err)
		require.Equal(t, string(wantTmpl), string(gotTmpl))
	})
}
