package templates

import (
	"fmt"
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplates_GetManifest(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := NewTemplates(c.Config)

	gotTmpl, err := tmpl.GetManifest()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("./templates/create/porter.yaml")
	assert.Equal(t, string(wantTmpl), string(gotTmpl))
}

func TestTemplates_GetRunScript(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := NewTemplates(c.Config)

	gotTmpl, err := tmpl.GetRunScript()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("./templates/build/cnab/app/run")
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestTemplates_GetDockerfile(t *testing.T) {
	testcases := []string{"buildkit", "docker"}
	for _, driver := range testcases {
		c := config.NewTestConfig(t)
		tmpl := NewTemplates(c.Config)

		gotTmpl, err := tmpl.GetDockerfile(driver)
		require.NoError(t, err)

		wantTmpl, _ := ioutil.ReadFile(fmt.Sprintf("./templates/build/%s.Dockerfile", driver))
		assert.Equal(t, wantTmpl, gotTmpl)
	}
}
