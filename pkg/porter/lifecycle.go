package porter

import (
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"github.com/pkg/errors"
)

// BundleAction is an interface that defines a method for supplying
// BundleLifecycleOptions.  This is useful when implementations contain
// action-specific options beyond the stock BundleLifecycleOptions.
type BundleAction interface {
	// GetAction returns the type of action: install, upgrade, invoke, uninstall
	GetAction() string

	// GetActionVerb returns the appropriate verb (present participle, e.g. -ing)
	// for the action.
	GetActionVerb() string

	GetOptions() *BundleActionOptions
}

type BundleActionOptions struct {
	sharedOptions
	BundlePullOptions
	AllowAccessToDockerHost bool
}

func (o *BundleActionOptions) Validate(args []string, porter *Porter) error {
	if o.Tag != "" {
		// Ignore anything set based on the bundle directory we are in, go off of the tag
		o.File = ""
		o.CNABFile = ""
		o.TagSet = true

		if err := o.validateTag(); err != nil {
			return err
		}
	}

	return o.sharedOptions.Validate(args, porter)
}

func (o *BundleActionOptions) GetOptions() *BundleActionOptions {
	return o
}

// BuildActionArgs converts an instance of user-provided action options into prepared arguments
// that can be used to execute the action.
func (p *Porter) BuildActionArgs(action BundleAction) (cnabprovider.ActionArguments, error) {
	opts := action.GetOptions()
	args := cnabprovider.ActionArguments{
		Action:                action.GetAction(),
		Installation:          opts.Name,
		BundlePath:            opts.CNABFile,
		Params:                make(map[string]string, len(opts.combinedParameters)),
		CredentialIdentifiers: make([]string, len(opts.CredentialIdentifiers)),
		Driver:                opts.Driver,
		RelocationMapping:     opts.RelocationMapping,
		AllowDockerHostAccess: opts.AllowAccessToDockerHost,
	}

	err := opts.LoadParameters(p)
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	// Do a safe copy so that modifications to the args aren't also made to the
	// original options, which is confusing to debug
	for k, v := range opts.combinedParameters {
		args.Params[k] = v
	}
	copy(args.CredentialIdentifiers, opts.CredentialIdentifiers)

	return args, nil
}

// prepullBundleByTag handles calling the bundle pull operation and updating
// the shared options like name and bundle file path. This is used by install, upgrade
// and uninstall
func (p *Porter) prepullBundleByTag(opts *BundleActionOptions) error {
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
