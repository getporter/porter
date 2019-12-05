package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_PrintMixins(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	opts := PrintMixinsOptions{
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatTable,
		},
	}
	err := p.PrintMixins(opts)

	require.Nil(t, err)
	wantOutput := `Name   Version   Author
exec   v1.0      Deis Labs
`
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, wantOutput, gotOutput)
}

func TestPorter_InstallMixin(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	opts := mixin.InstallOptions{
		Name: "exec",
		URL:  "https://example.com",
	}
	err := p.InstallMixin(opts)

	require.NoError(t, err)

	wantOutput := "installed exec mixin to ~/.porter/mixins/exec\nexec mixin v1.0 (abc123)"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotOutput)
}
