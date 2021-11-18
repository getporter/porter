package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/build/buildkit"
	"get.porter.sh/porter/pkg/build/docker"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"github.com/stretchr/testify/assert"
)

func TestPorter_GetBuilder(t *testing.T) {
	t.Run("docker", func(t *testing.T) {
		p := Porter{Config: &config.Config{}}
		p.SetExperimentalFlags(experimental.FlagBuildDrivers)
		p.Data.BuildDriver = config.BuildDriverDocker
		driver := p.GetBuilder()
		assert.IsType(t, &docker.Builder{}, driver)
	})
	t.Run("buildkit", func(t *testing.T) {
		p := Porter{Config: &config.Config{}}
		p.SetExperimentalFlags(experimental.FlagBuildDrivers)
		p.Data.BuildDriver = config.BuildDriverBuildkit
		driver := p.GetBuilder()
		assert.IsType(t, &buildkit.Builder{}, driver)
	})
	t.Run("unspecified", func(t *testing.T) {
		// Always default to Docker
		p := Porter{Config: &config.Config{}}
		p.SetExperimentalFlags(experimental.FlagBuildDrivers)
		p.Data.BuildDriver = ""
		driver := p.GetBuilder()
		assert.IsType(t, &docker.Builder{}, driver)
	})
	t.Run("buildkit - experimental flag disabled", func(t *testing.T) {
		// Default to docker when the experimental flag isn't set
		p := Porter{Config: &config.Config{}}
		p.Data.BuildDriver = "buildkit"
		driver := p.GetBuilder()
		assert.IsType(t, &docker.Builder{}, driver)
	})
}
