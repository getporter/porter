package cnabprovider

import (
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"

	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/cnab-go/driver"
	dockerdriver "github.com/deislabs/cnab-go/driver/docker"
	"github.com/deislabs/cnab-go/driver/lookup"

	"github.com/deislabs/porter/pkg/config"
)

type Duffle struct {
	*config.Config
}

func NewDuffle(c *config.Config) *Duffle {
	return &Duffle{
		Config: c,
	}
}

func (d *Duffle) newDriver(driverName, claimName string) (driver.Driver, error) {
	driverImpl, err := lookup.Lookup(driverName)
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

	// Setup host mount for outputs
	err = d.setupOutputsMount(driverImpl, claimName)
	if err != nil {
		return nil, err
	}

	return driverImpl, err
}

func (d *Duffle) setupOutputsMount(driverImpl driver.Driver, claimName string) error {
	// If docker driver, setup host bind mount for outputs
	if dockerish, ok := driverImpl.(*dockerdriver.Driver); ok {
		outputsDir, err := d.Config.GetOutputsDir()
		if err != nil {
			return errors.Wrap(err, "unable to get outputs directory")
		}

		// Create source outputs sub-directory using the bundle name
		bundleOutputsDir := filepath.Join(outputsDir, claimName)
		err = d.FileSystem.MkdirAll(bundleOutputsDir, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "could not create outputs directory %s for docker driver bind mount", bundleOutputsDir)
		}

		var cfgOpt dockerdriver.ConfigurationOption = func(containerCfg *container.Config, hostCfg *container.HostConfig) error {
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
	return nil
}

// WriteClaimOutputs writes outputs to a claim, according to the provided bundle
// and Duffle config
func (d *Duffle) WriteClaimOutputs(c *claim.Claim, action string) error {
	if c.Bundle == nil {
		return errors.New("claim has no bundle")
	}

	if c.Bundle.Outputs == nil {
		return nil
	}

	for name, output := range c.Bundle.Outputs {
		// TODO: refactor with cnab-go logic: https://github.com/deislabs/cnab-go/pull/99
		if config.OutputAppliesTo(action, output) {
			output, err := d.ReadBundleOutput(name, c.Name)
			if err != nil {
				return errors.Wrapf(err, "unable to read output %s", name)
			}
			c.Outputs[name] = output.Value
		}
	}
	return nil
}
