package cnabprovider

import (
	"fmt"

	"get.porter.sh/porter/pkg/manifest"
	"github.com/deislabs/cnab-go/action"
	"github.com/deislabs/cnab-go/claim"
	"github.com/pkg/errors"
)

func (d *Runtime) Uninstall(args ActionArguments) error {
	c, err := d.instanceStorage.Read(args.Claim)
	if err != nil {
		// Yay! It's already gone
		if err == claim.ErrClaimNotFound {
			return nil
		}
		return errors.Wrapf(err, "could not load bundle instance %s", args.Claim)
	}

	if args.BundlePath != "" {
		// TODO: if they installed an insecure bundle, do they really need to do --insecure again to uninstall it?
		c.Bundle, err = d.LoadBundle(args.BundlePath, args.Insecure)
		if err != nil {
			return err
		}
	}

	c.Parameters, err = d.loadParameters(&c, args.Params, string(manifest.ActionUninstall))
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}

	driver, err := d.newDriver(args.Driver, c.Name, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}
	i := action.Uninstall{
		Driver: driver,
	}

	creds, err := d.loadCredentials(c.Bundle, args.CredentialIdentifiers)
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
		paramKeys := make([]string, 0, len(c.Parameters))
		for k := range c.Parameters {
			paramKeys = append(paramKeys, k)
		}
		fmt.Fprintf(d.Err, "uninstalling bundle %s (%s) as %s\n\tparams: %v\n\tcreds: %v\n", c.Bundle.Name, args.BundlePath, c.Name, paramKeys, credKeys)
	}

	err = i.Run(&c, creds, d.ApplyConfig(args)...)
	if err != nil {
		return errors.Wrap(err, "failed to uninstall the bundle")
	}

	err = d.instanceStorage.Delete(args.Claim)

	return errors.Wrap(err, "failed to remove the record of the bundle")
}
