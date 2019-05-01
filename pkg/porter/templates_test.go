package porter

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplates_GetManifestTemplate(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	gotTmpl, err := p.GetManifestTemplate()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("./templates/porter.yaml")
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestTemplates_GetRunScriptTemplate(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	gotTmpl, err := p.GetRunScriptTemplate()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("./templates/run")
	assert.Equal(t, wantTmpl, gotTmpl)
}

func TestTemplates_GetDockerfile(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	gotTmpl, err := p.GetDockerfile()
	require.NoError(t, err)

	wantTmpl, _ := ioutil.ReadFile("./templates/Dockerfile")
	assert.Equal(t, wantTmpl, gotTmpl)
}
