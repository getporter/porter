package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/build/buildkit"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/tests"
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

func TestPorter_LoadWithSecretResolveError(t *testing.T) {
	ctx := context.Background()
	p := NewTestPorter(t)

	// Use a config file that has a secret, we aren't setting the secret value so resolve will fail
	p.TestConfig.TestContext.AddTestFileFromRoot("tests/testdata/config/config-with-storage-secret.yaml", "/home/myuser/.porter/config.yaml")

	// Configure porter to read the config file
	p.TestConfig.DataLoader = config.LoadFromEnvironment()

	err := p.Connect(ctx)

	// Validate the porter is handling the error
	tests.RequireErrorContains(t, err, "secret not found")
}
