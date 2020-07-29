package porter

import (
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"github.com/pkg/errors"
)

type BundleLifecycleOpts struct {
	sharedOptions
	BundlePullOptions
	AllowAccessToDockerHost bool
}

func (o *BundleLifecycleOpts) Validate(args []string, porter *Porter) error {
	err := o.sharedOptions.Validate(args, porter)
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
		Action:                deperator.Action,
		Installation:          o.Name,
		BundlePath:            o.CNABFile,
		Params:                make(map[string]string, len(o.combinedParameters)),
		CredentialIdentifiers: make([]string, len(o.CredentialIdentifiers)),
		Driver:                o.Driver,
		RelocationMapping:     o.RelocationMapping,
		AllowDockerHostAccess: o.AllowAccessToDockerHost,
	}

	// Do a safe copy so that modifications to the args aren't also made to the
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

	cachedBundle, err := p.PullBundle(opts.BundlePullOptions)
	if err != nil {
		return errors.Wrapf(err, "unable to pull bundle %s", opts.Tag)
	}

	opts.CNABFile = cachedBundle.BundlePath
	opts.RelocationMapping = cachedBundle.RelocationFilePath

	if opts.Name == "" {
		opts.Name = cachedBundle.Bundle.Name
	}

	if cachedBundle.Manifest != nil {
		p.Manifest = cachedBundle.Manifest
	}

	return nil
}
