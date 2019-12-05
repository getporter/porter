package cnabprovider

import (
	"os"

	"get.porter.sh/porter/pkg/config"
	instancestorage "get.porter.sh/porter/pkg/instance-storage"
	"github.com/deislabs/cnab-go/driver"
	"github.com/deislabs/cnab-go/driver/lookup"
)

type Runtime struct {
	*config.Config
	instanceStorage instancestorage.StorageProvider
}

func NewRuntime(c *config.Config, instanceStorage instancestorage.StorageProvider) *Runtime {
	return &Runtime{
		Config:          c,
		instanceStorage: instanceStorage,
	}
}

func (d *Runtime) newDriver(driverName string, claimName string, args ActionArguments) (driver.Driver, error) {
	driverImpl, err := lookup.Lookup(driverName)
	if err != nil {
		return driverImpl, err
	}

	if configurable, ok := driverImpl.(driver.Configurable); ok {
		driverCfg := make(map[string]string)
		// Load any driver-specific config out of the environment
		for env := range configurable.Config() {
			if val, ok := os.LookupEnv(env); ok {
				driverCfg[env] = val
			}
		}

		configurable.SetConfig(driverCfg)
	}

	return driverImpl, err
}
