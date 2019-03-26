package cnabprovider

import (
	"os"

	"github.com/deislabs/porter/pkg/config"

	"github.com/deislabs/duffle/pkg/driver"
)

type Duffle struct {
	*config.Config
}

func NewDuffle(c *config.Config) *Duffle {
	return &Duffle{
		Config: c,
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
