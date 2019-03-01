package porter

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixin_PrintSchema(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	err := p.PrintManifestSchema()
	require.NoError(t, err)

	gotSchema := p.TestConfig.TestContext.GetOutput()

	wantSchema, err := ioutil.ReadFile("testdata/schema.json")
	require.NoError(t, err)

	assert.Equal(t, string(wantSchema), gotSchema)
}
