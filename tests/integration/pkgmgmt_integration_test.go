//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageManager_GetSchema(t *testing.T) {
	ctx := context.Background()

	c := config.NewTestConfig(t)
	c.TestContext.UseFilesystem()

	// bin is my home now
	binDir := c.TestContext.FindBinDir()
	c.SetHomeDir(binDir)

	p := mixin.NewPackageManager(c.Config)
	gotSchema, err := p.GetSchema(ctx, "exec")
	require.NoError(t, err)

	wantSchema, err := os.ReadFile("../../pkg/exec/schema/exec.json")
	require.NoError(t, err)
	assert.Equal(t, string(wantSchema), gotSchema)
}
