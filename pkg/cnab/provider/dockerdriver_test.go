package cnabprovider

import (
	"testing"

	instancestorageprovider "github.com/deislabs/porter/pkg/instance-storage/provider"

	"github.com/stretchr/testify/require"

	"github.com/deislabs/cnab-go/driver/docker"

	"github.com/deislabs/porter/pkg/config"
)

func TestNewDriver_Docker(t *testing.T) {
	c := config.NewTestConfig(t)
	instanceStorage := instancestorageprovider.NewPluginDelegator(c.Config)
	d := NewRuntime(c.Config, instanceStorage)

	driver, err := d.newDriver("docker", "myclaim", ActionArguments{})
	require.NoError(t, err)

	if _, ok := driver.(*docker.Driver); ok {
		// TODO: check dockerConfigurationOptions to verify expected bind mount setup,
		// once we're able to (add ability to dockerdriver pkg)
	} else {
		t.Fatal("expected driver to be of type *dockerdriver.Driver")
	}
}
