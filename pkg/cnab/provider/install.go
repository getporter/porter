package cnabprovider

import (
	"fmt"

	"github.com/deislabs/cnab-go/action"
	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/porter/pkg/config"
	"github.com/pkg/errors"
)

func (d *Duffle) Install(args ActionArguments) error {
	// TODO: this entire function should be exposed in a duffle sdk package e.g. duffle.Install
	// we shouldn't be reimplementing calling all these functions all over again

	c, err := claim.New(args.Claim)
	if err != nil {
		return errors.Wrap(err, "invalid bundle instance name")
	}

	b, err := d.LoadBundle(args.BundlePath, args.Insecure)
	if err != nil {
		return err
	}

	err = b.Validate()
	if err != nil {
		return errors.Wrap(err, "invalid bundle")
	}
	c.Bundle = b

	params, err := d.loadParameters(c, args.Params, string(config.ActionInstall))
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}
	c.Parameters = params

	dvr, err := d.newDriver(args.Driver, c.Name, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}
	i := action.Install{
		Driver: dvr,
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
		fmt.Fprintf(d.Err, "installing bundle %s (%s) as %s\n\tparams: %v\n\tcreds: %v\n", c.Bundle.Name, args.BundlePath, c.Name, paramKeys, credKeys)
	}

	// Install and capture error
	runErr := i.Run(c, creds, d.ApplyConfig(args)...)

	// Add/update the outputs section of a claim and capture error
	writeErr := d.WriteClaimOutputs(c, string(config.ActionInstall))

	// ALWAYS write out a claim, even if the installation fails
	claimStore, err := d.NewClaimStore()
	if err != nil {
		return errors.Wrap(err, "could not access claim store")
	}
	saveErr := claimStore.Store(*c)
	if runErr != nil {
		return errors.Wrap(runErr, "failed to install the bundle")
	}
	if writeErr != nil {
		return errors.Wrap(writeErr, "failed to write outputs to the bundle instance")
	}
	return errors.Wrap(saveErr, "failed to record the installation for the bundle")
}
