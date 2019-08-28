package cnabprovider

import (
	"github.com/deislabs/cnab-go/driver"
)

// Shared arguments for all CNAB actions supported by duffle
type ActionArguments struct {
	// Name of the instance.
	Claim string

	// Either a filepath to the bundle or the name of the bundle.
	BundlePath string

	// Additional files to copy into the bundle
	// Target Path => File Contents
	Files map[string]string

	// Insecure bundle action allowed.
	Insecure bool

	// Params is the set of parameters to pass to the bundle.
	Params map[string]string

	// Either a filepath to a credential file or the name of a set of a credentials.
	CredentialIdentifiers []string

	// Driver is the CNAB-compliant driver used to run bundle actions.
	Driver string
}

func (args ActionArguments) ApplyFiles() func(op *driver.Operation) error {
	return func(op *driver.Operation) error {
		for k, v := range args.Files {
			op.Files[k] = v
		}
		return nil
	}
}
