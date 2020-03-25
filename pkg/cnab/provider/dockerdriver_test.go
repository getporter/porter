package cnabprovider

import (
	"testing"

	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/stretchr/testify/require"
)

func TestNewDriver_Docker(t *testing.T) {
	d := NewTestRuntime(t)

	driver, err := d.newDriver(DriverNameDocker, "myclaim", ActionArguments{})
	require.NoError(t, err)

	if _, ok := driver.(*docker.Driver); ok {
		// TODO: check dockerConfigurationOptions to verify expected bind mount setup,
		// once we're able to (add ability to dockerdriver pkg)
	} else {
		t.Fatal("expected driver to be of type *dockerdriver.Driver")
	}
}
