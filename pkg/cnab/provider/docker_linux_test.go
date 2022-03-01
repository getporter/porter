package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_getDockerGroupID(t *testing.T) {
	cfg := config.NewTestConfig(t)
	cfg.Setenv(test.ExpectedCommandEnv, "getent group docker")
	cfg.Setenv(test.ExpectedCommandOutputEnv, "docker:x:103")

	r := NewRuntime(cfg.Config, nil, nil)
	gid, err := r.getDockerGroupId()
	require.NoError(t, err)
	assert.Equal(t, "103", gid)
}
