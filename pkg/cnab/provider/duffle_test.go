package cnabprovider

import (
	"testing"

	"github.com/stretchr/testify/require"

	duffledriver "github.com/deislabs/duffle/pkg/driver"
	"github.com/deislabs/porter/pkg/config"
)

func TestNewDriver_Docker(t *testing.T) {
	c := config.NewTestConfig(t)
	d := NewDuffle(c.Config)

	driver, err := d.newDriver("docker", "myclaim")
	require.NoError(t, err)

	if _, ok := driver.(*duffledriver.DockerDriver); ok {
		// TODO: check dockerConfigurationOptions to verify expected bind mount setup,
		// once we're able to (add ability to duffledriver pkg)
	} else {
		t.Fatal("expected driver to be of type *duffledriver.DockerDriver")
	}
}
