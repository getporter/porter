package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/encoding"
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

	bundleRef *cnab.BundleReference
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
	// Some actions need to resolve this early
	if opts.bundleRef != nil {
		return *opts.bundleRef, nil
	}

	var bundleRef cnab.BundleReference

	useReference := func(ref cnab.OCIReference) error {
		pullOpts := *opts // make a copy just to do the pull
		pullOpts.Reference = ref.String()
		cachedBundle, err := p.prepullBundleByReference(&pullOpts)
		if err != nil {
			return errors.Wrapf(err, "unable to pull bundle %s", opts.Reference)
		}

		bundleRef = cachedBundle.BundleReference
		return nil
	}

	// load the referenced bundle
	if opts.Reference != "" {
		if err := useReference(opts.GetReference()); err != nil {
			return cnab.BundleReference{}, err
		}
	} else if opts.File != "" { // load the local bundle source
		localBundle, err := p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
		if err != nil {
			return cnab.BundleReference{}, err
		}
		bundleRef = localBundle
	} else if opts.CNABFile != "" { // load the cnab bundle definition
		bun, err := p.CNAB.LoadBundle(opts.CNABFile)
		if err != nil {
			return cnab.BundleReference{}, err
		}
		bundleRef = cnab.BundleReference{Definition: bun}
	} else if opts.Name != "" { // Return the bundle associated with the installation
		i, err := p.Claims.GetInstallation(opts.Namespace, opts.Name)
		if err != nil {
			return cnab.BundleReference{}, errors.Wrapf(err, "installation %s/%s not found", opts.Namespace, opts.Name)
		}
		if i.Status.BundleReference != "" {
			ref, err := cnab.ParseOCIReference(i.Status.BundleReference)
			if err != nil {
				return cnab.BundleReference{}, errors.Wrapf(err, "installation.Status.BundleReference is invalid")
			}
			if err := useReference(ref); err != nil {
				return cnab.BundleReference{}, err
			}
		} else { // The bundle was installed from source
			lastRun, err := p.Claims.GetLastRun(opts.Namespace, opts.Name)
			if err != nil {
				return cnab.BundleReference{}, errors.Wrap(err, "could not load the bundle definition from the installation's last run")
			}

			bundleRef = cnab.BundleReference{
				Definition: cnab.ExtendedBundle{lastRun.Bundle},
				Digest:     digest.Digest(lastRun.BundleDigest)}

			if lastRun.BundleReference != "" {
				bundleRef.Reference, err = cnab.ParseOCIReference(lastRun.BundleReference)
				if err != nil {
					return cnab.BundleReference{}, errors.Wrapf(err, "invalid bundle reference, %s, found on the last bundle run record %s", lastRun.BundleReference, lastRun.ID)
				}
			}
		}
	} else { // Nothing was referenced
		return cnab.BundleReference{}, errors.New("No bundle specified")
	}

	if opts.Name == "" {
		opts.Name = bundleRef.Definition.Name
	}

	opts.bundleRef = &bundleRef
	return bundleRef, nil
}

// BuildActionArgs converts an instance of user-provided action options into prepared arguments
// that can be used to execute the action.
func (p *Porter) BuildActionArgs(installation claims.Installation, action BundleAction) (cnabprovider.ActionArguments, error) {
	opts := action.GetOptions()
	bundleRef, err := p.resolveBundleReference(opts)

	if opts.RelocationMapping != "" {
		err := encoding.UnmarshalFile(p.FileSystem, opts.RelocationMapping, &bundleRef.RelocationMap)
		if err != nil {
			return cnabprovider.ActionArguments{}, errors.Wrapf(err, "could not parse the relocation mapping file at %s", opts.RelocationMapping)
		}
	}

	// Resolve the final set of typed parameters, taking into account the user overrides, parameter sources
	// and defaults
	err = opts.LoadParameters(p)
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	if p.Debug {
		fmt.Fprintln(p.Err, "resolving parameters for installation", installation)
	}
	resolvedParams, err := p.resolveParameters(installation, bundleRef.Definition, action.GetAction(), opts.combinedParameters)
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	args := cnabprovider.ActionArguments{
		Action:                action.GetAction(),
		Installation:          installation,
		BundleReference:       bundleRef,
		Params:                resolvedParams,
		CredentialIdentifiers: make([]string, len(opts.CredentialIdentifiers)),
		Driver:                opts.Driver,
		AllowDockerHostAccess: opts.AllowAccessToDockerHost,
	}

	// Do a safe copy so that modifications to the args aren't also made to the
	// original options, which is confusing to debug
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

	opts.RelocationMapping = cachedBundle.RelocationFilePath

	if opts.Name == "" {
		opts.Name = cachedBundle.Definition.Name
	}

	if cachedBundle.Manifest != nil {
		p.Manifest = cachedBundle.Manifest
	}

	return cachedBundle, nil
}
