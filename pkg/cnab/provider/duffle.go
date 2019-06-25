package cnabprovider

import (
	"os"
	"path/filepath"

	"github.com/deislabs/cnab-go/driver"
	duffledriver "github.com/deislabs/duffle/pkg/driver"
	"github.com/deislabs/porter/pkg/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"
)

type Duffle struct {
	*config.Config
}

func NewDuffle(c *config.Config) *Duffle {
	return &Duffle{
		Config: c,
	}
}

func (d *Duffle) newDriver(driverName string) (driver.Driver, error) {
	driverImpl, err := duffledriver.Lookup(driverName)
	if err != nil {
		return driverImpl, err
	}

	// Load any driver-specific config out of the environment.
	// TODO: This should be exposed in duffle, taken from cmd/duffle/main.go prepareDriver
	if configurable, ok := driverImpl.(driver.Configurable); ok {
		driverCfg := map[string]string{}
		for env := range configurable.Config() {
			driverCfg[env] = os.Getenv(env)
		}
		configurable.SetConfig(driverCfg)
	}

	// If docker driver, setup host bind mount for outputs
	if dockerish, ok := driverImpl.(*duffledriver.DockerDriver); ok {
		outputsDir, err := d.Config.GetOutputsDir()
		if err != nil {
			return nil, errors.Wrap(err, "unable to get outputs directory")
		}

		// Create outputs sub-directory using the manifest name, if it does not already exist
		bundleOutputsDir := filepath.Join(outputsDir, d.Manifest.Name)
		err = d.FileSystem.MkdirAll(bundleOutputsDir, 0755)
		if err != nil {
			return nil, errors.Wrapf(err, "could not create outputs directory %s for docker driver bind mount", bundleOutputsDir)
		}

		var cfgOpt duffledriver.DockerConfigurationOption = func(containerCfg *container.Config, hostCfg *container.HostConfig) error {
			outputsMount := mount.Mount{
				Type:   mount.TypeBind,
				Source: bundleOutputsDir,
				Target: config.BundleOutputsDir,
			}
			hostCfg.Mounts = append(hostCfg.Mounts, outputsMount)
			return nil
		}
		dockerish.AddConfigurationOptions(cfgOpt)
	}

	return driverImpl, err
}
