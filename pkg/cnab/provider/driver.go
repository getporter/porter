package cnabprovider

import (
	"get.porter.sh/porter/pkg/cnab/drivers"
	"get.porter.sh/porter/pkg/cnab/extensions"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"
)

const (
	// DriverNameDocker is the name of the CNAB Docker driver.
	DriverNameDocker = "docker"

	// DriverNameDebug is the name of the CNAB debug driver.
	DriverNameDebug = "debug"
)

func (r *Runtime) newDriver(driverName string, args ActionArguments) (driver.Driver, error) {
	var driverImpl driver.Driver
	var err error

	// Pull applicable extension from list of processed extensions
	dockerExt, dockerRequired, err := r.Extensions.GetDocker()
	if err != nil {
		return nil, err
	}

	if args.AllowDockerHostAccess {
		if driverName != DriverNameDocker {
			return nil, errors.Errorf("allow-docker-host-access was enabled, but the driver is %s", driverName)
		}

		driverImpl, err = r.dockerDriverWithHostAccess(dockerExt)
	} else {
		if dockerRequired {
			return nil, errors.Errorf("extension %q is required but allow-docker-host-access was not enabled",
				extensions.DockerExtensionKey)
		}
		driverImpl, err = drivers.LookupDriver(r.Context, driverName)
	}
	if err != nil {
		return nil, err
	}

	if configurable, ok := driverImpl.(driver.Configurable); ok {
		driverCfg := make(map[string]string)
		// Load any driver-specific config out of the environment
		for env := range configurable.Config() {
			if val, ok := r.LookupEnv(env); ok {
				driverCfg[env] = val
			}
		}

		configurable.SetConfig(driverCfg)
	}

	return driverImpl, nil
}

func (r *Runtime) dockerDriverWithHostAccess(config extensions.Docker) (driver.Driver, error) {
	const dockerSock = "/var/run/docker.sock"

	if exists, _ := r.FileSystem.Exists(dockerSock); !exists {
		return nil, errors.Errorf("allow-docker-host-access was specified but could not detect a local docker daemon running by checking for %s", dockerSock)
	}

	d := &docker.Driver{}
	d.AddConfigurationOptions(func(cfg *container.Config, hostCfg *container.HostConfig) error {

		// Equivalent of using: --privileged
		// Required for DinD, or "Docker-in-Docker"
		hostCfg.Privileged = config.Privileged

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
	return driver.Driver(d), nil
}
