package cnabprovider

import (
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/driver"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/hashicorp/go-multierror"
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

func (r *Runtime) ApplyConfig(args ActionArguments) action.OperationConfigs {
	return action.OperationConfigs{
		r.SetOutput(),
		r.AddFiles(args),
		r.AddRelocation(args),
	}
}

func (r *Runtime) SetOutput() action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		op.Out = r.Out
		return nil
	}
}

func (r *Runtime) AddFiles(args ActionArguments) action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		for k, v := range args.Files {
			op.Files[k] = v
		}

		// Add claim.json to file list as well, if exists
		claim, err := r.claims.ReadLastClaim(args.Installation)
		if err == nil {
			claimBytes, err := yaml.Marshal(claim)
			if err != nil {
				return errors.Wrapf(err, "could not marshal claim %s for installation %s", claim.ID, args.Installation)
			}
			op.Files[config.ClaimFilepath] = string(claimBytes)
		}

		return nil
	}
}

// AddRelocation operates on an ActionArguments and adds any provided relocation mapping
// to the operation's files.
func (r *Runtime) AddRelocation(args ActionArguments) action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		if args.RelocationMapping != "" {
			b, err := r.FileSystem.ReadFile(args.RelocationMapping)
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

// appendFailedResult creates a failed result from the operation error and accumulates
// the error(s).
func (r *Runtime) appendFailedResult(opErr error, c claim.Claim) error {
	saveResult := func() error {
		result, err := c.NewResult(claim.StatusFailed)
		if err != nil {
			return err
		}
		return r.claims.SaveResult(result)
	}

	resultErr := saveResult()

	// Accumulate any errors from the operation with the persistence errors
	return multierror.Append(opErr, resultErr).ErrorOrNil()
}

func (r *Runtime) printDebugInfo(creds valuesource.Set, params map[string]interface{}) {
	if r.Debug {
		// only print out the names of the credentials, not the contents, cuz they big and sekret
		credKeys := make([]string, 0, len(creds))
		for k := range creds {
			credKeys = append(credKeys, k)
		}
		// param values may also be sensitive, so just print names
		paramKeys := make([]string, 0, len(params))
		for k := range params {
			paramKeys = append(paramKeys, k)
		}
		fmt.Fprintf(r.Err, "params: %v\ncreds: %v\n", paramKeys, credKeys)
	}
}
