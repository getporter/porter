package cnabprovider

import (
	"fmt"

	"github.com/deislabs/cnab-go/action"
	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/config"
)

func (d *Duffle) Uninstall(args ActionArguments) error {
	// TODO: this entire function should be exposed in a duffle sdk package e.g. duffle.Install
	// we shouldn't be reimplementing calling all these functions all over again

	claims := d.NewClaimStore()
	claim, err := claims.Read(args.Claim)
	if err != nil {
		return errors.Wrapf(err, "could not load claim %s", args.Claim)
	}

	if args.BundleIdentifier != "" {
		// TODO: handle resolving based on bundle name
		// TODO: if they installed an insecure bundle, do they really need to do --insecure again to unisntall it?
		claim.Bundle, err = d.LoadBundle(args.BundleIdentifier, args.Insecure)
		if err != nil {
			return err
		}
	}

	if len(args.Params) > 0 {
		claim.Parameters, err = d.loadParameters(&claim, args.Params, string(config.ActionUninstall))
		if err != nil {
			return errors.Wrap(err, "invalid parameters")
		}
	}

	driver, err := d.newDriver(args.Driver, claim.Name)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}
	i := action.Uninstall{
		Driver: driver,
	}

	creds, err := d.loadCredentials(claim.Bundle, args.CredentialIdentifiers)
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
		paramKeys := make([]string, 0, len(claim.Parameters))
		for k := range claim.Parameters {
			paramKeys = append(paramKeys, k)
		}
		fmt.Fprintf(d.Err, "uninstalling bundle %s (%s) as %s\n\tparams: %v\n\tcreds: %v\n", claim.Bundle.Name, args.BundleIdentifier, claim.Name, paramKeys, credKeys)
	}

	err = i.Run(&claim, creds, d.Out)
	if err != nil {
		return errors.Wrap(err, "failed to uninstall the bundle")
	}

	err = claims.Delete(args.Claim)
	if err != nil {
		return errors.Wrap(err, "failed to remove the record of the bundle")
	}

	return nil
}
