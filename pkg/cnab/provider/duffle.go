package cnabprovider

import (
	"os"

	"github.com/deislabs/duffle/pkg/driver"
	"github.com/deislabs/porter/pkg/context"
)

type Duffle struct {
	*context.Context
}

func NewDuffle(c *context.Context) *Duffle {
	return &Duffle{
		Context: c,
	}
}

func (d *Duffle) newDockerDriver() *driver.DockerDriver {
	dd := &driver.DockerDriver{}

	// Load any driver-specific config out of the environment.
	// TODO: This should be exposed in duffle, taken from cmd/duffle/main.go prepareDriver
	driverCfg := map[string]string{}
	for env := range dd.Config() {
		driverCfg[env] = os.Getenv(env)
	}
	dd.SetConfig(driverCfg)

	return dd
}
