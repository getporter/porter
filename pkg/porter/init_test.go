package porter

import (
	"testing"

	"github.com/deislabs/porter/pkg/config"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	p, _ := NewTestPorter(t)
	InitializePorterHome(t, p)

	err := p.Init()
	require.NoError(t, err)

	configFileExists, err := p.FileSystem.Exists(config.Name)
	require.NoError(t, err)
	assert.True(t, configFileExists)
}
