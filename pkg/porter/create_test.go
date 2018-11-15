package porter

import (
	"testing"

	"github.com/deislabs/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	err := p.Create()
	require.NoError(t, err)

	configFileExists, err := p.FileSystem.Exists(config.Name)
	require.NoError(t, err)
	assert.True(t, configFileExists)

	runScriptExists, err := p.FileSystem.Exists(config.RunScript)
	require.NoError(t, err)
	assert.True(t, runScriptExists)
}
