package cnabprovider

import (
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/cnab/drivers"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/docker/docker/api/types/container"
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

	if args.AllowDockerHostAccess && driverName != DriverNameDocker {
		return nil, fmt.Errorf("allow-docker-host-access was enabled, but the driver is %s", driverName)
	}

	if dockerRequired && !args.AllowDockerHostAccess {
		return nil, fmt.Errorf("extension %q is required but allow-docker-host-access was not enabled", cnab.DockerExtensionKey)
	}

	if len(args.HostVolumeMounts) > 0 && driverName != DriverNameDocker {
		return nil, fmt.Errorf("mount-host-volume was was used to mount a volume, but the driver is %s", driverName)
	}

	if !args.AllowDockerHostAccess && len(args.HostVolumeMounts) == 0 {
		driverImpl, err = drivers.LookupDriver(r.Context, driverName)
	}

	var d *docker.Driver
	if args.AllowDockerHostAccess || len(args.HostVolumeMounts) > 0 {
		d = &docker.Driver{}
	}

	if args.AllowDockerHostAccess {
		driverImpl, err = r.dockerDriverWithHostAccess(dockerExt, d)
	}

	if len(args.HostVolumeMounts) > 0 {
		driverImpl, err = func(dr *docker.Driver) (driver.Driver, error) {

			dr.AddConfigurationOptions(func(cfg *container.Config, hostCfg *container.HostConfig) error {
				err := r.addVolumeMountsToHostConfig(hostCfg, args.HostVolumeMounts)
				if err != nil {
					return err
				}
				return nil
			})

			return driver.Driver(dr), nil
		}(d)
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

func (r *Runtime) dockerDriverWithHostAccess(config cnab.Docker, d *docker.Driver) (driver.Driver, error) {

	// Run the container with privileged access if necessary
	if config.Privileged {
		d.AddConfigurationOptions(func(cfg *container.Config, hostCfg *container.HostConfig) error {
			// Equivalent of using: --privileged
			// Required for DinD, or "Docker-in-Docker"
			hostCfg.Privileged = true
			return nil
		})
	}

	// Mount the docker socket
	d.AddConfigurationOptions(r.mountDockerSocket)

	return driver.Driver(d), nil
}

func (r *Runtime) addVolumeMountsToHostConfig(hostConfig *container.HostConfig, mounts []HostVolumeMountSpec) error {
	for _, mount := range mounts {
		err := r.addVolumeMountToHostConfig(hostConfig, mount.Source, mount.Target, mount.ReadOnly)
		if err != nil {
			return err
		}
	}
	return nil
}
