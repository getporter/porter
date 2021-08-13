package porter

import (
	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"github.com/opencontainers/go-digest"
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
	var err error

	if o.Reference != "" {
		// Ignore anything set based on the bundle directory we are in, go off of the tag
		o.File = ""
		o.CNABFile = ""
		o.ReferenceSet = true

		if err := o.BundlePullOptions.Validate(); err != nil {
			return err
		}
	}

	err = o.sharedOptions.Validate(args, porter)
	if err != nil {
		return err
	}

	if o.Name == "" && o.File == "" && o.CNABFile == "" && o.Reference == "" {
		return errors.New("No bundle specified. Either an installation name, --reference, --file or --cnab-file must be specified or the current directory must contain a porter.yaml file.")
	}

	return nil
}

func (o *BundleActionOptions) GetOptions() *BundleActionOptions {
	return o
}

func (p *Porter) resolveBundleReference(opts *BundleActionOptions) (cnab.BundleReference, error) {
	// Return the referenced bundle
	if opts.Reference != "" {
		cachedBundle, err := p.prepullBundleByReference(opts)
		if err != nil {
			return cnab.BundleReference{}, errors.Wrapf(err, "unable to pull bundle %s", opts.Reference)
		}
		return cachedBundle.BundleReference, nil
	}

	// Return the local bundle source
	if opts.File != "" {
		// Return the local bundle source
		err := p.applyDefaultOptions(&opts.sharedOptions)
		if err != nil {
			return cnab.BundleReference{}, err
		}
		return p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	}

	// Return the cnab bundle
	if opts.CNABFile != "" {
		bun, err := p.CNAB.LoadBundle(opts.CNABFile)
		if err != nil {
			return cnab.BundleReference{}, err
		}
		return cnab.BundleReference{Definition: bun}, nil
	}

	// Return the bundle associated with the installation
	if opts.Name != "" {
		lastRun, err := p.Claims.GetLastRun(opts.Namespace, opts.Name)
		if err != nil {
			return cnab.BundleReference{}, errors.Wrap(err, "could not load the bundle definition from the installation's last run")
		}

		bundleRef := cnab.BundleReference{
			Definition: lastRun.Bundle,
			Digest:     digest.Digest(lastRun.BundleDigest)}

		if lastRun.BundleReference != "" {
			bundleRef.Reference, err = cnab.ParseOCIReference(lastRun.BundleReference)
			if err != nil {
				return cnab.BundleReference{}, errors.Wrapf(err, "invalid bundle reference, %s, found on the last bundle run record %s", lastRun.BundleReference, lastRun.ID)
			}
		}

		return bundleRef, nil
	}

	// Nothing was referenced
	return cnab.BundleReference{}, errors.New("No bundle specified")
}

// BuildActionArgs converts an instance of user-provided action options into prepared arguments
// that can be used to execute the action.
func (p *Porter) BuildActionArgs(action BundleAction) (cnabprovider.ActionArguments, error) {
	opts := action.GetOptions()
	args := cnabprovider.ActionArguments{
		// TODO(carolynvs): set the bundle digest
		BundleReference:       opts.Reference,
		Action:                action.GetAction(),
		Installation:          opts.Name,
		Namespace:             opts.Namespace,
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

// prepullBundleByReference handles calling the bundle pull operation and updating
// the shared options like name and bundle file path. This is used by install, upgrade
// and uninstall
func (p *Porter) prepullBundleByReference(opts *BundleActionOptions) (cache.CachedBundle, error) {
	if opts.Reference == "" {
		return cache.CachedBundle{}, nil
	}

	cachedBundle, err := p.PullBundle(opts.BundlePullOptions)
	if err != nil {
		return cache.CachedBundle{}, errors.Wrapf(err, "unable to pull bundle %s", opts.Reference)
	}

	opts.CNABFile = cachedBundle.BundlePath
	opts.RelocationMapping = cachedBundle.RelocationFilePath

	if opts.Name == "" {
		opts.Name = cachedBundle.Definition.Name
	}

	if cachedBundle.Manifest != nil {
		p.Manifest = cachedBundle.Manifest
	}

	return cachedBundle, nil
}
