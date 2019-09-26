package porter

import (
	"github.com/deislabs/cnab-go/bundle"
	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

type BundleLifecycleOpts struct {
	sharedOptions
	BundlePullOptions
}

func (o *BundleLifecycleOpts) Validate(args []string, cxt *context.Context) error {
	err := o.sharedOptions.Validate(args, cxt)
	if err != nil {
		return err
	}
	if o.Tag != "" {
		// Ignore anything set based on the bundle directory we are in, go off of the tag
		o.File = ""
		o.CNABFile = ""

		return o.validateTag()
	}
	return nil
}

// ToActionArgs converts this instance of user-provided action options.
func (o *BundleLifecycleOpts) ToActionArgs(deperator *dependencyExecutioner) cnabprovider.ActionArguments {
	args := cnabprovider.ActionArguments{
		Claim:                 o.Name,
		BundlePath:            o.CNABFile,
		Insecure:              o.Insecure,
		Params:                make(map[string]string, len(o.combinedParameters)),
		CredentialIdentifiers: make([]string, len(o.CredentialIdentifiers)),
		Driver:                o.Driver,
	}

	// Do a safe copy so that modifications to the duffle args aren't also made to the
	// original options, which is confusing to debug
	for k, v := range o.combinedParameters {
		args.Params[k] = v
	}
	copy(args.CredentialIdentifiers, o.CredentialIdentifiers)

	deperator.ApplyDependencyMappings(&args)

	return args
}

// prepullBundleByTag handles calling the bundle pull operation and updating
// the shared options like name and bundle file path. This is used by install, upgrade
// and uninstall
func (p *Porter) prepullBundleByTag(opts *BundleLifecycleOpts) error {
	if opts.Tag == "" {
		return nil
	}

	bundlePath, err := p.PullBundle(opts.BundlePullOptions)
	if err != nil {
		return errors.Wrapf(err, "unable to pull bundle %s", opts.Tag)
	}
	opts.CNABFile = bundlePath
	rdr, err := p.Config.FileSystem.Open(bundlePath)
	if err != nil {
		return errors.Wrap(err, "unable to open bundle file")
	}
	defer rdr.Close()
	bun, err := bundle.ParseReader(rdr)
	if opts.Name == "" {
		opts.Name = bun.Name
	}
	return nil
}
