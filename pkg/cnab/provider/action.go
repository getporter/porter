package cnabprovider

import (
	"encoding/json"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Shared arguments for all CNAB actions
type ActionArguments struct {
	// Name of the installation.
	Installation string

	// Either a filepath to the bundle or the name of the bundle.
	BundlePath string

	// Additional files to copy into the bundle
	// Target Path => File Contents
	Files map[string]string

	// Params is the set of user-specified parameter values to pass to the bundle.
	Params map[string]string

	// ParameterSets is a list of strings representing either a filepath to a
	// parameter set file or the name of a set of a parameters.
	ParameterSets []string

	// Either a filepath to a credential file or the name of a set of a credentials.
	CredentialIdentifiers []string

	// Driver is the CNAB-compliant driver used to run bundle actions.
	Driver string

	// Path to an optional relocation mapping file
	RelocationMapping string

	// Give the bundle privileged access to the docker daemon.
	AllowDockerHostAccess bool
}

func (d *Runtime) ApplyConfig(args ActionArguments) action.OperationConfigs {
	return action.OperationConfigs{
		d.SetOutput(),
		d.AddFiles(args),
		d.AddRelocation(args),
	}
}

func (d *Runtime) SetOutput() action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		op.Out = d.Out
		return nil
	}
}

func (d *Runtime) AddFiles(args ActionArguments) action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		for k, v := range args.Files {
			op.Files[k] = v
		}

		// Add claim.json to file list as well, if exists
		claimName := args.Installation
		claim, err := d.claims.Read(claimName)
		if err == nil {
			claimBytes, err := yaml.Marshal(claim)
			if err != nil {
				return errors.Wrapf(err, "could not marshal claim %s", claimName)
			}
			op.Files[config.ClaimFilepath] = string(claimBytes)
		}

		return nil
	}
}

// AddRelocation operates on an ActionArguments and adds any provided relocation mapping
// to the operation's files.
func (d *Runtime) AddRelocation(args ActionArguments) action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		if args.RelocationMapping != "" {
			b, err := d.FileSystem.ReadFile(args.RelocationMapping)
			if err != nil {
				return errors.Wrap(err, "unable to add relocation mapping")
			}
			op.Files["/cnab/app/relocation-mapping.json"] = string(b)
			var reloMap relocation.ImageRelocationMap
			err = json.Unmarshal(b, &reloMap)
			// If the invocation image is present in the relocation mapping, we need
			// to update the operation and set the new image reference. Unfortunately,
			// the relocation mapping is just reference => reference, so there isn't a
			// great way to check for the invocation image.
			if mappedInvo, ok := reloMap[op.Image.Image]; ok {
				op.Image.Image = mappedInvo
			}
		}
		return nil
	}
}
