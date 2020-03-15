package cnabprovider

import (
	"os"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/credentials"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/cnabio/cnab-go/driver/lookup"
	"github.com/docker/docker/api/types/container"
)

type Runtime struct {
	*config.Config
	credentials credentials.CredentialProvider
	claims      claims.ClaimProvider
}

func NewRuntime(c *config.Config, claims claims.ClaimProvider, credentials credentials.CredentialProvider) *Runtime {
	return &Runtime{
		Config:      c,
		claims:      claims,
		credentials: credentials,
	}
}

func (d *Runtime) newDriver(driverName string, claimName string, args ActionArguments) (driver.Driver, error) {
	var driverImpl driver.Driver

	// TODO: Remove once this PR is merged: https://github.com/cnabio/cnab-go/pull/199
	if driverName == "docker" {
		dockerDriver := &docker.Driver{}
		if val, ok := os.LookupEnv("PRIVILEGED"); ok && val == "1" {
			dockerDriver.AddConfigurationOptions(func(cfg *container.Config, hostCfg *container.HostConfig) error {
				hostCfg.Privileged = true
				return nil
			})
		}
		driverImpl = driver.Driver(dockerDriver)
	} else {
		driverImpl, err := lookup.Lookup(driverName)
		if err != nil {
			return driverImpl, err
		}
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

	return driverImpl, nil
}
