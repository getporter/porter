package porter

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/printer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_PrintMixins(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	opts := PrintMixinsOptions{
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatTable,
		},
	}
	err := p.PrintMixins(opts)

	require.Nil(t, err)
	wantOutput := `Name   Version   Author
exec   v1.0      Porter Authors
`
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, wantOutput, gotOutput)
}

func TestPorter_InstallMixin(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	opts := mixin.InstallOptions{}
	opts.Name = "exec"
	opts.URL = "https://example.com"

	err := p.InstallMixin(opts)

	require.NoError(t, err)

	wantOutput := "installed exec mixin v1.0 (abc123)\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotOutput)
}

func TestPorter_UninstallMixin(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	opts := pkgmgmt.UninstallOptions{}
	err := opts.Validate([]string{"exec"})
	require.NoError(t, err, "Validate failed")

	err = p.UninstallMixin(opts)
	require.NoError(t, err, "UninstallMixin failed")

	wantOutput := "Uninstalled exec mixin"
	gotoutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotoutput)
}

func TestPorter_CreateMixin(t *testing.T) {
	p := NewTestPorter(t)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	porterTempDir := homeDir + "/temp"

	err = os.Mkdir(porterTempDir, 0755)
	require.NoError(t, err)

	opts := MixinsCreateOptions{
		MixinName:      "MyMixin",
		AuthorName:     "Author Name",
		AuthorUsername: "username",
		DirPath:        porterTempDir,
	}

	err = p.CreateMixin(opts)
	require.NoError(t, err)

	wantOutput := `Created MyMixin mixin`
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotOutput)

	err = os.RemoveAll(porterTempDir)
	require.NoError(t, err)
}
