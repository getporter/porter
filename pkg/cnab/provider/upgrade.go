package cnabprovider

import (
	"fmt"

	"get.porter.sh/porter/pkg/cnab/extensions"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/cnabio/cnab-go/action"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func (d *Runtime) Upgrade(args ActionArguments) error {
	c, err := d.claims.Read(args.Installation)
	if err != nil {
		return errors.Wrapf(err, "could not load installation %s", args.Installation)
	}

	if args.BundlePath != "" {
		c.Bundle, err = d.LoadBundle(args.BundlePath)
		if err != nil {
			return err
		}
	}

	exts, err := extensions.ProcessRequiredExtensions(c.Bundle)
	if err != nil {
		return errors.Wrap(err, "unable to process required extensions")
	}
	d.Extensions = exts

	c.Parameters, err = d.loadParameters(&c, args.Params, args.ParameterSets, string(manifest.ActionUpgrade))
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}

	driver, err := d.newDriver(args.Driver, c.Installation, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}
	i := action.Upgrade{
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
		fmt.Fprintf(d.Err, "upgrading bundle %s (%s) as %s\n\tparams: %v\n\tcreds: %v\n", c.Bundle.Name, args.BundlePath, c.Installation, paramKeys, credKeys)
	}

	var result *multierror.Error
	// Upgrade and capture error
	err = i.Run(&c, creds, d.ApplyConfig(args)...)
	if err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to upgrade the bundle"))
	}

	// ALWAYS write out a claim, even if the upgrade fails
	err = d.claims.Save(c)
	if err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to record the upgrade for the bundle"))
	}

	return result.ErrorOrNil()
}
