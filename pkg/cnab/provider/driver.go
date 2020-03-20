package cnabprovider

import (
	"os"

	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/cnabio/cnab-go/driver/lookup"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"
)

const (
	// DriverNameDocker is the name of the CNAB Docker driver.
	DriverNameDocker = "docker"

	// DriverNameDocker is the name of the CNAB debug driver.
	DriverNameDebug = "debug"
)

func (d *Runtime) newDriver(driverName string, claimName string, args ActionArguments) (driver.Driver, error) {
	var driverImpl driver.Driver

	allowDockerHostAccess, err := d.Data.GetAllowDockerHostAccess()
	if err != nil {
		return nil, err
	}

	if allowDockerHostAccess {
		if driverName != DriverNameDocker {
			return nil, errors.Wrapf(err, "Docker host access was enabled, but the driver is %s.", driverName)
		}
		driverImpl = dockerDriverWithHostAccess()
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

func dockerDriverWithHostAccess() driver.Driver {
	d := &docker.Driver{}
	d.AddConfigurationOptions(func(cfg *container.Config, hostCfg *container.HostConfig) error {

		// Equivalent of using: --privileged
		// Required for DinD, or "Docker-in-Docker"
		hostCfg.Privileged = true

		// Equivalent of using: -v /var/run/docker.sock:/var/run/docker.sock
		// Required for DooD, or "Docker-out-of-Docker"
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

		return nil
	})
	return driver.Driver(d)
}
