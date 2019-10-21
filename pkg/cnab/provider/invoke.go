package cnabprovider

import (
	"fmt"

	cnabaction "github.com/deislabs/cnab-go/action"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func (d *Runtime) Invoke(action string, args ActionArguments) error {
	c, err := d.instanceStorage.Read(args.Claim)
	if err != nil {
		return errors.Wrapf(err, "could not load claim %s", args.Claim)
	}

	if args.BundlePath != "" {
		c.Bundle, err = d.LoadBundle(args.BundlePath, args.Insecure)
		if err != nil {
			return err
		}
	}

	if len(args.Params) > 0 {
		c.Parameters, err = d.loadParameters(&c, args.Params, action)
		if err != nil {
			return errors.Wrap(err, "invalid parameters")
		}
	}

	driver, err := d.newDriver(args.Driver, c.Name, args)
	if err != nil {
		return errors.Wrap(err, "unable to instantiate driver")
	}

	i := cnabaction.RunCustom{
		Action: action,
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
		fmt.Fprintf(d.Err, "invoking bundle %s (%s) with action %s as %s\n\tparams: %v\n\tcreds: %v\n", c.Bundle.Name, args.BundlePath, action, c.Name, paramKeys, credKeys)
	}

	var result *multierror.Error
	// Run the action and ALWAYS write out a claim, even if the action fails
	err = i.Run(&c, creds, d.ApplyConfig(args)...)
	if err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to invoke the bundle"))
	}

	// ALWAYS write out a claim, even if the action fails
	err = d.instanceStorage.Store(c)
	if err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to record the updated claim for the bundle"))
	}

	return result.ErrorOrNil()
}
