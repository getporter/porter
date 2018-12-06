package porter

import (
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_readOutputs(t *testing.T) {
	p := NewTestPorter(t)

	p.TestConfig.TestContext.AddTestFile("testdata/outputs1.txt", filepath.Join(mixin.OutputsDir, "myoutput1"))
	p.TestConfig.TestContext.AddTestFile("testdata/outputs2.txt", filepath.Join(mixin.OutputsDir, "myoutput2"))
	gotOutputs, err := p.readOutputs()
	require.NoError(t, err)

	wantOutputs := []string{
		"FOO=BAR",
		"BAZ=QUX",
		"A=B",
	}
	assert.Equal(t, wantOutputs, gotOutputs)
}
