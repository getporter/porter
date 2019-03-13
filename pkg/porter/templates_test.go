package porter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
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
