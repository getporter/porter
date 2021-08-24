package cnabprovider

import (
	"os"

	"get.porter.sh/porter/pkg/cnab/drivers"
	"get.porter.sh/porter/pkg/cnab/extensions"
	"github.com/cnabio/cnab-go/bundle"
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

	if driverName == "docker" && r.shouldPullInvocationImage(args.BundleReference.Definition){
		// Always ensure that the local docker cache has the repository digests for the invocation image
		os.Setenv("PULL_ALWAYS", "1")
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

func (r *Runtime) shouldPullInvocationImage(b bundle.Bundle) bool {
	// When the invocation image is expected to have a repository digest, we should always try to pull it
	for _, ii := range b.InvocationImages {
		if ii.Digest != "" {
			return true
		}
	}
	return false
}

func (r *Runtime) dockerDriverWithHostAccess(config extensions.Docker) (driver.Driver, error) {
	const dockerSock = "/var/run/docker.sock"

	if exists, _ := r.FileSystem.Exists(dockerSock); !exists {
		return nil, errors.Errorf("allow-docker-host-access was specified but could not detect a local docker daemon running by checking for %s", dockerSock)
	}

	d, err := drivers.LookupDriver(r.Context, "docker")
	if err != nil {
		return nil, err
	}

	dockerDriver, ok := d.(*docker.Driver)
	if !ok {
		return nil, errors.New("could not create a Docker driver")
	}
	dockerDriver.AddConfigurationOptions(func(cfg *container.Config, hostCfg *container.HostConfig) error {
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
	return d, nil
}
