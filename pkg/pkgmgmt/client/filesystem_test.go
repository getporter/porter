package client

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystem_List(t *testing.T) {
	c := config.NewTestConfig(t)

	p := NewFileSystem(c.Config, "mixins")
	mixins, err := p.List()

	require.Nil(t, err)
	require.Len(t, mixins, 2)
	assert.Equal(t, mixins[0], "exec")
	assert.Equal(t, mixins[1], "testmixin")
}
