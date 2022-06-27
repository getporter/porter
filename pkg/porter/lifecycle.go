package porter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/cnab"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/storage"
	"github.com/opencontainers/go-digest"
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

	// GetOptions returns the common bundle action options used to execute the bundle.
	GetOptions() *BundleActionOptions

	// Validate the action before it is executed.
	Validate(ctx context.Context, args []string, p *Porter) error
}

type BundleActionOptions struct {
	sharedOptions
	BundlePullOptions
	AllowDockerHostAccess bool
	NoLogs                bool

	bundleRef *cnab.BundleReference
}

func (o *BundleActionOptions) Validate(ctx context.Context, args []string, porter *Porter) error {
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

	// When you run porter installation apply, there are some settings from porter install
	// that aren't exposed as flags. This allows the user to set them in the config file
	// and we will use them before running the bundle.
	if o.Driver == "" {
		// We have both porter build --driver, and porter install --driver
		// So in the config file it's named build-driver and runtime-driver
		// This is why we check first before applying the value. Only apply the config
		// file setting if they didn't specify a flag.
		o.Driver = porter.Config.Data.RuntimeDriver
	}
	o.AllowDockerHostAccess = porter.Config.Data.AllowDockerHostAccess

	err = o.sharedOptions.Validate(ctx, args, porter)
	if err != nil {
		return err
	}

	if o.Name == "" && o.File == "" && o.CNABFile == "" && o.Reference == "" {
		return errors.New("no bundle specified. Either an installation name, --reference, --file or --cnab-file must be specified or the current directory must contain a porter.yaml file")
	}

	return nil
}

func (o *BundleActionOptions) GetOptions() *BundleActionOptions {
	return o
}

func (p *Porter) resolveBundleReference(ctx context.Context, opts *BundleActionOptions) (cnab.BundleReference, error) {
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
			return err
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
		localBundle, err := p.ensureLocalBundleIsUpToDate(ctx, opts.bundleFileOptions)
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
		i, err := p.Installations.GetInstallation(ctx, opts.Namespace, opts.Name)
		if err != nil {
			return cnab.BundleReference{}, fmt.Errorf("installation %s/%s not found: %w", opts.Namespace, opts.Name, err)
		}
		if i.Status.BundleReference != "" {
			ref, err := cnab.ParseOCIReference(i.Status.BundleReference)
			if err != nil {
				return cnab.BundleReference{}, fmt.Errorf("installation.Status.BundleReference is invalid: %w", err)
			}
			if err := useReference(ref); err != nil {
				return cnab.BundleReference{}, err
			}
		} else { // The bundle was installed from source
			lastRun, err := p.Installations.GetLastRun(ctx, opts.Namespace, opts.Name)
			if err != nil {
				return cnab.BundleReference{}, fmt.Errorf("could not load the bundle definition from the installation's last run: %w", err)
			}

			bundleRef = cnab.BundleReference{
				Definition: cnab.NewBundle(lastRun.Bundle),
				Digest:     digest.Digest(lastRun.BundleDigest)}

			if lastRun.BundleReference != "" {
				bundleRef.Reference, err = cnab.ParseOCIReference(lastRun.BundleReference)
				if err != nil {
					return cnab.BundleReference{}, fmt.Errorf("invalid bundle reference, %s, found on the last bundle run record %s: %w", lastRun.BundleReference, lastRun.ID, err)
				}
			}
		}
	} else { // Nothing was referenced
		return cnab.BundleReference{}, errors.New("no bundle specified")
	}

	if opts.Name == "" {
		opts.Name = bundleRef.Definition.Name
	}

	opts.bundleRef = &bundleRef
	return bundleRef, nil
}

// BuildActionArgs converts an instance of user-provided action options into prepared arguments
// that can be used to execute the action.
func (p *Porter) BuildActionArgs(ctx context.Context, installation storage.Installation, action BundleAction) (cnabprovider.ActionArguments, error) {
	opts := action.GetOptions()
	bundleRef, err := p.resolveBundleReference(ctx, opts)
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	if opts.RelocationMapping != "" {
		err := encoding.UnmarshalFile(p.FileSystem, opts.RelocationMapping, &bundleRef.RelocationMap)
		if err != nil {
			return cnabprovider.ActionArguments{}, fmt.Errorf("could not parse the relocation mapping file at %s: %w", opts.RelocationMapping, err)
		}
	}

	// Resolve the final set of typed parameters, taking into account the user overrides, parameter sources
	// and defaults
	err = opts.LoadParameters(ctx, p, opts.bundleRef.Definition)
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	if p.Debug {
		fmt.Fprintln(p.Err, "resolving parameters for installation", installation)
	}

	// Do not resolve parameters from dependencies
	params := make(map[string]string, len(opts.combinedParameters))
	for k, v := range opts.combinedParameters {
		if strings.Contains(k, "#") {
			continue
		}
		params[k] = v
	}
	resolvedParams, err := p.resolveParameters(ctx, installation, bundleRef.Definition, action.GetAction(), params)
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	args := cnabprovider.ActionArguments{
		Action:                action.GetAction(),
		Installation:          installation,
		BundleReference:       bundleRef,
		Params:                resolvedParams,
		Driver:                opts.Driver,
		AllowDockerHostAccess: opts.AllowDockerHostAccess,
		PersistLogs:           !opts.NoLogs,
	}

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
		return cache.CachedBundle{}, err
	}

	opts.RelocationMapping = cachedBundle.RelocationFilePath

	if opts.Name == "" {
		opts.Name = cachedBundle.Definition.Name
	}

	return cachedBundle, nil
}
