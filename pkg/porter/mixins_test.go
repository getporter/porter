package porter

import (
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/mixin"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_PrintMixins(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	// Start with a fresh mixins dir to make the test more durable as we add more mixins later
	mixinsDir, err := p.GetMixinsDir()
	require.Nil(t, err)
	err = p.FileSystem.RemoveAll(mixinsDir)
	require.Nil(t, err)

	// Just copy in the exec and helm mixins
	srcMixinsDir := filepath.Join(p.TestConfig.TestContext.FindBinDir(), "mixins")
	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(srcMixinsDir, "/helm"), filepath.Join(mixinsDir, "helm"))
	p.TestConfig.TestContext.AddTestDirectory(filepath.Join(srcMixinsDir, "/exec"), filepath.Join(mixinsDir, "exec"))

	opts := PrintMixinsOptions{Format: printer.FormatTable}
	err = p.PrintMixins(opts)

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
