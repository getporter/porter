package porter

import (
	"context"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_PrintMixins(t *testing.T) {
	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	opts := PrintMixinsOptions{
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatPlaintext,
		},
	}
	err := p.PrintMixins(ctx, opts)

	require.Nil(t, err)
	gotOutput := p.TestConfig.TestContext.GetOutput()
	test.CompareGoldenFile(t, "mixins/list-output.txt", gotOutput)
}

func TestPorter_InstallMixin(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := mixin.InstallOptions{}
	opts.Name = "exec"
	opts.URL = "https://example.com"

	err := p.InstallMixin(context.Background(), opts)

	require.NoError(t, err)

	wantOutput := "installed exec mixin v1.0 (abc123)\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotOutput)
}

func TestPorter_UninstallMixin(t *testing.T) {
	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	opts := pkgmgmt.UninstallOptions{}
	err := opts.Validate([]string{"exec"})
	require.NoError(t, err, "Validate failed")

	err = p.UninstallMixin(ctx, opts)
	require.NoError(t, err, "UninstallMixin failed")

	wantOutput := "Uninstalled exec mixin"
	gotoutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotoutput)
}

func TestPorter_CreateMixin(t *testing.T) {
	p := NewTestPorter(t)

	tempDir, err := p.FileSystem.TempDir("", "porter")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	opts := MixinsCreateOptions{
		MixinName:      "MyMixin",
		AuthorName:     "Author Name",
		AuthorUsername: "username",
		DirPath:        tempDir,
	}

	err = p.CreateMixin(opts)
	require.NoError(t, err)

	wantOutput := "Created MyMixin mixin\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotOutput)
}
