package porter

import (
	"fmt"

	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/credentials"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

func (p *Porter) PrintCredentials(opts printer.PrintOptions) error {
	return nil
}

type CredentialOptions struct {
	sharedOptions
	DryRun bool
	Silent bool
}

// Validate prepares for an action and validates the options.
// For example, relative paths are converted to full paths and then checked that
// they exist and are accessible.
func (g *CredentialOptions) Validate(args []string, cxt *context.Context) error {
	err := g.validateCredName(args)
	if err != nil {
		return err
	}

	err = g.defaultBundleFiles(cxt)
	if err != nil {
		return err
	}

	return g.validateBundleJson(cxt)
}

func (g *CredentialOptions) validateCredName(args []string) error {
	if len(args) == 1 {
		g.Name = args[0]
	} else if len(args) > 1 {
		return errors.Errorf("only one positional argument may be specified, the credential name, but multiple were received: %s", args)
	}
	return nil
}

func (p *Porter) GenerateCredentials(opts CredentialOptions) error {

	//TODO make this work for either porter.yaml OR a bundle
	bundle, err := p.CNAB.LoadBundle(opts.CNABFile, opts.Insecure)
	if err != nil {
		return err
	}
	name := opts.Name
	if name == "" {
		name = bundle.Name
	}
	genOpts := credentials.GenerateOptions{
		Name:        name,
		Credentials: bundle.Credentials,
		Silent:      opts.Silent,
	}
	fmt.Fprintf(p.Out, "Generating new credential %s from bundle %s\n", genOpts.Name, bundle.Name)
	fmt.Fprintf(p.Out, "==> %d credentials required for bundle %s\n", len(genOpts.Credentials), bundle.Name)

	cs, err := credentials.GenerateCredentials(genOpts)
	if err != nil {
		return errors.Wrap(err, "unable to generate credentials")
	}

	//write the credential out to PORTER_HOME with Porter's Context
	data, err := yaml.Marshal(cs)
	if err != nil {
		return errors.Wrap(err, "unable to generate credentials YAML")
	}
	if opts.DryRun {
		fmt.Fprintf(p.Out, "%v", string(data))
		return nil
	}

	dest, err := p.Config.GetCredentialPath(genOpts.Name)
	if err != nil {
		return errors.Wrap(err, "unable to determine credentials directory")
	}

	fmt.Fprintf(p.Out, "Saving credential to %s\n", dest)
	err = p.Context.FileSystem.WriteFile(dest, data, 0600)
	if err != nil {
		return errors.Wrapf(err, "couldn't write credential file %s", dest)
	}
	return nil
}
