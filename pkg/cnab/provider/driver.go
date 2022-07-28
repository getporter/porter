package cnabprovider

import (
	"os"
	"strings"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/cnab/drivers"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
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
			return nil, fmt.Errorf("allow-docker-host-access was enabled, but the driver is %s", driverName)
		}

		driverImpl, err = r.dockerDriverWithHostAccess(dockerExt)
	} else {
		if dockerRequired {
			return nil, fmt.Errorf("extension %q is required but allow-docker-host-access was not enabled",
				cnab.DockerExtensionKey)
		}
		driverImpl, err = drivers.LookupDriver(r.Context, driverName)
	}
	if err != nil {
		return nil, err
	}

	// Handle Directory support for the docker driver
	// TODO: Handle directory support for other runtimes -- how?
	if driverName == "docker" && r.Extensions.DirectoryParameterSupport() {
		d := driverImpl.(*docker.Driver)
		for _, _dd := range r.Extensions[cnab.DirectoryParameterExtension.Key].([]cnab.DirectoryDetails) {
				switch _dd.Kind {
				case cnab.ParameterSourceTypeMount:
					// preserve the closure context by running in an immediate closure
					// AddConfigurationOptions executes its closure asynchronously and so only picks up
					// The last value for dd unless we do it this way
					func () {
						dd := _dd
						d.AddConfigurationOptions(func(cfg *container.Config, hostCfg *container.HostConfig) error {
							x := dd.Mount
							x.Type = "bind"
							x.ReadOnly = !dd.Writeable
							pairs := make([]string, len(os.Environ())*2)
							for i, env := range os.Environ() {
								parts := strings.Split(env, "=")
								pairs[i*2] = "$" + parts[0]
								pairs[i*2+1] = parts[1]
							}
	
							rep := strings.NewReplacer(pairs...)
							x.Source = rep.Replace(x.Source)
							x.Target = rep.Replace(x.Target)
							if hostCfg.Mounts == nil || len(hostCfg.Mounts) < 1 {
								hostCfg.Mounts = []mount.Mount{
									x.Mount,
								}
							} else {
								hostCfg.Mounts = append(hostCfg.Mounts, x.Mount)
							}
							return nil
						})

					}()
				}
			}
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

func (r *Runtime) dockerDriverWithHostAccess(config cnab.Docker) (driver.Driver, error) {
	const dockerSock = "/var/run/docker.sock"

	if exists, _ := r.FileSystem.Exists(dockerSock); !exists {
		return nil, fmt.Errorf("allow-docker-host-access was specified but could not detect a local docker daemon running by checking for %s", dockerSock)
	}

	d := &docker.Driver{}

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
