package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/build/buildkit"
	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestPorter_GetBuilder(t *testing.T) {
	// Now that docker is deprecated, always use buildkit for now
	// I didn't remove the config capabilities in case we need it later

	t.Run("docker deprecated, use buildkit", func(t *testing.T) {
		p := Porter{Config: &config.Config{}}
		p.Data.BuildDriver = config.BuildDriverDocker
		driver := p.GetBuilder(context.Background())
		assert.IsType(t, &buildkit.Builder{}, driver)
	})
	t.Run("buildkit", func(t *testing.T) {
		p := Porter{Config: &config.Config{}}
		p.Data.BuildDriver = config.BuildDriverBuildkit
		driver := p.GetBuilder(context.Background())
		assert.IsType(t, &buildkit.Builder{}, driver)
	})
	t.Run("unspecified", func(t *testing.T) {
		// Always default to Docker
		p := Porter{Config: &config.Config{}}
		p.Data.BuildDriver = ""
		driver := p.GetBuilder(context.Background())
		assert.IsType(t, &buildkit.Builder{}, driver)
	})
}
