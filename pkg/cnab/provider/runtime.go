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
	"github.com/docker/docker/api/types/mount"
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

	if d.Dind && driverName == "docker" {
		driverImpl = dockerDindDriver()
	} else {
		var err error
		driverImpl, err = lookup.Lookup(driverName)
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

// If Dind support is enabled ("Docker-in-Docker"), set
// the equivalent of the following "docker run" options:
// 1.) -v /var/run/docker.sock:/var/run/docker.sock
// 2.) --privileged
func dockerDindDriver() driver.Driver {
	dockerDriver := &docker.Driver{}
	dockerDriver.AddConfigurationOptions(func(cfg *container.Config, hostCfg *container.HostConfig) error {
		dockerSockMount := mount.Mount{
			Source:   "/var/run/docker.sock",
			Target:   "/var/run/docker.sock",
			Type:     "bind",
			ReadOnly: false,
		}
		if hostCfg.Mounts == nil {
			hostCfg.Mounts = []mount.Mount{}
		}
		hostCfg.Mounts = append(hostCfg.Mounts, dockerSockMount)
		hostCfg.Privileged = true
		return nil
	})
	return driver.Driver(dockerDriver)
}
