package cnabprovider

import (
	"os"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-go/driver/lookup"
	instancestorage "github.com/deislabs/porter/pkg/storage"
)

type Runtime struct {
	*config.Config
	storage storage.StorageProvider
}

func NewRuntime(c *config.Config, s storage.StorageProvider) *Runtime {
	return &Runtime{
		Config:  c,
		storage: s,
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
