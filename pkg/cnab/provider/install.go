package cnabprovider

import (
	"fmt"

	"github.com/deislabs/duffle/pkg/action"
	"github.com/deislabs/duffle/pkg/claim"
	"github.com/pkg/errors"
)

type InstallArguments struct {
	// Name of the claim.
	Claim string

	// Either a filepath to the bundle or the name of the bundle.
	BundleIdentifier string

	// BundleIdentifier is a filepath.
	BundleIsFile bool

	// Insecure bundle installation allowed.
	Insecure bool

	// Params is the set of parameters to pass to the bundle.
	Params map[string]string

	// Either a filepath to a credential file or the name of a set of a credentials.
	CredentialIdentifiers []string
}

func (d *Duffle) Install(args InstallArguments) error {
	// TODO: this entire function should be exposed in a duffle sdk package e.g. duffle.Install
	// we shouldn't be reimplementing calling all these functions all over again

	c, err := claim.New(args.Claim)
	if err != nil {
		return errors.Wrap(err, "invalid claim name")
	}

	// TODO: handle resolving based on bundle name
	b, err := d.LoadBundle(args.BundleIdentifier, args.Insecure)
	if err != nil {
		return err
	}

	err = b.Validate()
	if err != nil {
		return errors.Wrap(err, "invalid bundle")
	}
	c.Bundle = b

	params, err := d.loadParameters(b, args.Params)
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}
	c.Parameters = params

	i := action.Install{
		Driver: d.newDockerDriver(),
	}
	creds, err := d.loadCredentials(b, args.CredentialIdentifiers)
	if err != nil {
		return errors.Wrap(err, "could not load credentials")
	}

	if d.Debug {
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
		fmt.Fprintf(d.Err, "installing bundle %s (%s) as %s\n\tparams: %v\n\tcreds: %v\n", c.Bundle.Name, args.BundleIdentifier, c.Name, paramKeys, credKeys)
	}

	// Install and ALWAYS write out a claim, even if the installation fails
	err = i.Run(c, creds, d.Out)
	saveErr := d.NewClaimStore().Store(*c)
	if err != nil {
		return errors.Wrap(err, "failed to install the bundle")
	}
	return errors.Wrap(saveErr, "failed to record the installation for the bundle")
}
