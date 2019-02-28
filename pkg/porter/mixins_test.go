package porter

import (
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/printer"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestPorter_GetMixins(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	// Start with a fresh mixins dir to make the test more durable as we add more mixins later
	mixinsDir, err := p.GetMixinsDir()
	require.Nil(t, err)
	err = p.FileSystem.RemoveAll(mixinsDir)
	require.Nil(t, err)

	// Just copy in the exec and helm mixins
	p.TestConfig.TestContext.AddTestDirectory("../../bin/mixins/helm", filepath.Join(mixinsDir, "helm"))
	p.TestConfig.TestContext.AddTestDirectory("../../bin/mixins/exec", filepath.Join(mixinsDir, "exec"))

	mixins, err := p.GetMixins()

	require.Nil(t, err)
	require.Len(t, mixins, 2)
	assert.Equal(t, mixins[0].Name, "exec")
	assert.Equal(t, mixins[1].Name, "helm")
}

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

	opts := printer.PrintOptions{Format: printer.FormatTable}
	err = p.PrintMixins(opts)

	require.Nil(t, err)
	wantOutput := "exec\nhelm\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotOutput)
}
