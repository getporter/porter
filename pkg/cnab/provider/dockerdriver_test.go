package cnabprovider

import (
	"testing"

	"github.com/stretchr/testify/require"

	dockerdriver "github.com/deislabs/cnab-go/driver/docker"
	"github.com/deislabs/porter/pkg/config"
)

func TestNewDriver_Docker(t *testing.T) {
	c := config.NewTestConfig(t)
	d := NewDuffle(c.Config)

	driver, err := d.newDriver("docker", "myclaim")
	require.NoError(t, err)

	if _, ok := driver.(*dockerdriver.Driver); ok {
		// TODO: check dockerConfigurationOptions to verify expected bind mount setup,
		// once we're able to (add ability to dockerdriver pkg)
	} else {
		t.Fatal("expected driver to be of type *dockerdriver.Driver")
	}
}
