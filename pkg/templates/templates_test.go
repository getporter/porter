package templates

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplates_GetManifest(t *testing.T) {
	tmpl := NewTemplates()

	gotTmpl, err := tmpl.GetManifest()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("./templates/create/porter.yaml")
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestTemplates_GetRunScript(t *testing.T) {
	tmpl := NewTemplates()

	gotTmpl, err := tmpl.GetRunScript()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("./templates/build/cnab/app/run")
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestTemplates_GetDockerfile(t *testing.T) {
	tmpl := NewTemplates()

	gotTmpl, err := tmpl.GetDockerfile()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("./templates/build/Dockerfile")
	assert.Equal(t, wantTmpl, gotTmpl)
}
