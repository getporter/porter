package builder

import (
	"context"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestFileOutput struct {
	Name     string
	FilePath string
}

func (o TestFileOutput) GetName() string {
	return o.Name
}

func (o TestFileOutput) GetFilePath() string {
	return o.FilePath
}

func TestFilePathOutputs(t *testing.T) {
	ctx := context.Background()
	c := runtime.NewTestRuntimeConfig(t)

	step := TestStep{
		Outputs: []Output{
			TestFileOutput{Name: "config", FilePath: "config.txt"},
		},
	}

	wantCfg := "abc123"
	err := c.FileSystem.WriteFile("config.txt", []byte(wantCfg), pkg.FileModeWritable)
	require.NoError(t, err, "could not write config.txt")

	err = ProcessFileOutputs(ctx, c.RuntimeConfig, step)
	require.NoError(t, err, "ProcessFileOutputs should not return an error")

	f := filepath.Join(portercontext.MixinOutputsDir, "config")
	gotOutput, err := c.FileSystem.ReadFile(f)
	require.NoError(t, err, "could not read output file %s", f)

	assert.Equal(t, wantCfg, string(gotOutput))
}
