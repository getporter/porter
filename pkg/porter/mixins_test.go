package porter

import (
	"testing"

	"github.com/deislabs/porter/pkg/mixin"
	"github.com/deislabs/porter/pkg/printer"
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
	wantOutput := "exec\nhelm\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotOutput)
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
