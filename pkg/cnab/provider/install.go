package cnabprovider

import (
	"fmt"

	"github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/claim"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"get.porter.sh/porter/pkg/cnab/extensions"
	"get.porter.sh/porter/pkg/manifest"
)

func (d *Runtime) Install(args ActionArguments) error {
	c, err := claim.New(args.Claim)
	if err != nil {
		return errors.Wrap(err, "invalid bundle instance name")
	}

	b, err := d.LoadBundle(args.BundlePath)
	if err != nil {
		return err
	}

	err = b.Validate()
	if err != nil {
		return errors.Wrap(err, "invalid bundle")
	}
	c.Bundle = b

	exts, err := extensions.ProcessRequiredExtensions(b)
	if err != nil {
		return errors.Wrap(err, "unable to process required extensions")
	}
	d.Extensions = exts

	params, err := d.loadParameters(c, args.Params, string(manifest.ActionInstall))
	if err != nil {
		return errors.Wrap(err, "invalid parameters")
	}
	c.Parameters = params

	dvr, err := d.newDriver(args.Driver, c.Installation, args)
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
		fmt.Fprintf(d.Err, "installing bundle %s (%s) as %s\n\tparams: %v\n\tcreds: %v\n", c.Bundle.Name, args.BundlePath, c.Installation, paramKeys, credKeys)
	}

	var result *multierror.Error
	// Install and capture error
	err = i.Run(c, creds, d.ApplyConfig(args)...)
	if err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to install the bundle"))
	}

	// ALWAYS write out a claim, even if the installation fails
	err = d.claims.Save(*c)
	if err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to record the installation for the bundle"))
	}

	return result.ErrorOrNil()
}
