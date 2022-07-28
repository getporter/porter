package porter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/cache"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/cnab/drivers"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
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
	GetOptions() *BundleExecutionOptions

	// Validate the action before it is executed.
	Validate(ctx context.Context, args []string, p *Porter) error
}

// BundleExecutionOptions are common options for commands that run a bundle (install/upgrade/invoke/uninstall)
type BundleExecutionOptions struct {
	*BundleReferenceOptions

	// AllowDockerHostAccess grants the bundle access to the Docker socket.
	AllowDockerHostAccess bool

	// DebugMode indicates if the bundle should be run in debug mode.
	DebugMode bool

	// NoLogs runs the bundle without persisting any logs.
	NoLogs bool

	// Params is the unparsed list of NAME=VALUE parameters set on the command line.
	Params []string

	// ParameterSets is a list of parameter sets containing parameter sources
	ParameterSets []string

	// CredentialIdentifiers is a list of credential names or paths to make available to the bundle.
	CredentialIdentifiers []string

	// Driver is the CNAB-compliant driver used to run bundle actions.
	Driver string

	// parsedParams is the parsed set of parameters from Params.
	parsedParams map[string]string

	// parsedParamSets is the parsed set of parameter from ParameterSets
	parsedParamSets map[string]string

	// combinedParameters is parsedParams merged on top of parsedParamSets.
	combinedParameters map[string]string
}

func NewBundleExecutionOptions() *BundleExecutionOptions {
	return &BundleExecutionOptions{
		BundleReferenceOptions: &BundleReferenceOptions{},
	}
}

func (o *BundleExecutionOptions) GetOptions() *BundleExecutionOptions {
	return o
}

func (o *BundleExecutionOptions) Validate(ctx context.Context, args []string, p *Porter) error {
	if err := o.BundleReferenceOptions.Validate(ctx, args, p); err != nil {
		return err
	}

	// Only validate the syntax of the --param flags
	// We will validate the parameter sets later once we have the bundle loaded.
	if err := o.parseParams(); err != nil {
		return err
	}

	o.defaultDriver(p)
	if err := o.validateDriver(p.Context); err != nil {
		return err
	}

	return nil
}

// LoadParameters validates and resolves the parameters and sets. It must be
// called after porter has loaded the bundle definition.
func (o *BundleExecutionOptions) LoadParameters(ctx context.Context, p *Porter, bun cnab.ExtendedBundle) error {
	// This is called in multiple code paths, so exit early if
	// we have already loaded the parameters into combinedParameters
	if o.combinedParameters != nil {
		return nil
	}

	err := o.parseParams()
	if err != nil {
		return err
	}

	err = o.parseParamSets(ctx, p, bun)
	if err != nil {
		return err
	}

	o.combinedParameters = o.combineParameters(p.Context)

	return nil
}

// parsedParams parses the variable assignments in Params.
func (o *BundleExecutionOptions) parseParams() error {
	p, err := storage.ParseVariableAssignments(o.Params)
	if err != nil {
		return err
	}

	o.parsedParams = p
	return nil
}

func (o *BundleExecutionOptions) populateInternalParameterSet(ctx context.Context, p *Porter, bun cnab.ExtendedBundle, i *storage.Installation) error {
	strategies := make([]secrets.Strategy, 0, len(o.parsedParams))
	for name, value := range o.parsedParams {
		strategies = append(strategies, storage.ValueStrategy(name, value))
	}

	strategies, err := p.Sanitizer.CleanParameters(ctx, strategies, bun, i.ID)
	if err != nil {
		return err
	}

	if len(strategies) == 0 {
		// if no override is specified, clear out the old parameters on the
		// installation record
		i.Parameters.Parameters = nil
		return nil
	}

	i.Parameters = i.NewInternalParameterSet(strategies...)

	return nil
}

// parseParamSets parses the variable assignments in ParameterSets.
func (o *BundleExecutionOptions) parseParamSets(ctx context.Context, p *Porter, bun cnab.ExtendedBundle) error {
	if len(o.ParameterSets) > 0 {
		parsed, err := p.loadParameterSets(ctx, bun, o.Namespace, o.ParameterSets)
		if err != nil {
			return fmt.Errorf("unable to process provided parameter sets: %w", err)
		}
		o.parsedParamSets = parsed
	}
	return nil
}

// Combine the parameters into a single map
// The params set on the command line take precedence over the params set in
// parameter set files
// Anything set multiple times, is decided by "last one set wins"
func (o *BundleExecutionOptions) combineParameters(c *portercontext.Context) map[string]string {
	final := make(map[string]string)

	for k, v := range o.parsedParamSets {
		final[k] = v
	}

	for k, v := range o.parsedParams {
		final[k] = v
	}

	//
	// Default the porter-debug param to --debug
	//
	if o.DebugMode {
		final["porter-debug"] = "true"
	}

	return final
}

// defaultDriver supplies the default driver if none is specified
func (o *BundleExecutionOptions) defaultDriver(p *Porter) {
	//
	// When you run porter installation apply, there are some settings from porter install
	// that aren't exposed as flags (like driver and allow-docker-host-access).
	// This allows the user to set them in the config file and we will use them before running the bundle.
	//

	// Apply global config to the --driver flag
	if o.Driver == "" {
		// We have both porter build --driver, and porter install --driver
		// So in the config file it's named build-driver and runtime-driver
		// This is why we check first before applying the value. Only apply the config
		// file setting if they didn't specify a flag.
		o.Driver = p.Data.RuntimeDriver
	}

	// Apply global config to the --allow-docker-host-access flag
	if !o.AllowDockerHostAccess {
		// Only apply the config setting if they didn't specify the flag (i.e. it's porter installation apply which doesn't have that flag)
		o.AllowDockerHostAccess = p.Config.Data.AllowDockerHostAccess
	}
}

// validateDriver validates that the provided driver is supported by Porter
func (o *BundleExecutionOptions) validateDriver(cxt *portercontext.Context) error {
	_, err := drivers.LookupDriver(cxt, o.Driver)
	return err
}

// BundleReferenceOptions are the set of options available for commands that accept a bundle reference
type BundleReferenceOptions struct {
	installationOptions
	BundlePullOptions

	bundleRef *cnab.BundleReference
}

func (o *BundleReferenceOptions) Validate(ctx context.Context, args []string, porter *Porter) error {
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

	err = o.installationOptions.Validate(ctx, args, porter)
	if err != nil {
		return err
	}

	if o.Name == "" && o.File == "" && o.CNABFile == "" && o.Reference == "" {
		return errors.New("no bundle specified. Either an installation name, --reference, --file or --cnab-file must be specified or the current directory must contain a porter.yaml file")
	}

	return nil
}

func (p *Porter) resolveBundleReference(ctx context.Context, opts *BundleReferenceOptions) (cnab.BundleReference, error) {
	// Some actions need to resolve this early
	if opts.bundleRef != nil {
		return *opts.bundleRef, nil
	}

	var bundleRef cnab.BundleReference

	useReference := func(ref cnab.OCIReference) error {
		pullOpts := *opts // make a copy just to do the pull
		pullOpts.Reference = ref.String()
		cachedBundle, err := p.prepullBundleByReference(ctx, &pullOpts)
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
	log := tracing.LoggerFromContext(ctx)

	opts := action.GetOptions()
	bundleRef, err := p.resolveBundleReference(ctx, opts.BundleReferenceOptions)
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	if opts.RelocationMapping != "" {
		err := encoding.UnmarshalFile(p.FileSystem, opts.RelocationMapping, &bundleRef.RelocationMap)
		if err != nil {
			return cnabprovider.ActionArguments{}, log.Error(fmt.Errorf("could not parse the relocation mapping file at %s: %w", opts.RelocationMapping, err))
		}
	}

	// Resolve the final set of typed parameters, taking into account the user overrides, parameter sources
	// and defaults
	err = opts.LoadParameters(ctx, p, opts.bundleRef.Definition)
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	log.Debugf("resolving parameters for installation %s", installation)

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
		return cnabprovider.ActionArguments{}, log.Error(err)
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
func (p *Porter) prepullBundleByReference(ctx context.Context, opts *BundleReferenceOptions) (cache.CachedBundle, error) {
	if opts.Reference == "" {
		return cache.CachedBundle{}, nil
	}

	cachedBundle, err := p.PullBundle(ctx, opts.BundlePullOptions)
	if err != nil {
		return cache.CachedBundle{}, err
	}

	opts.RelocationMapping = cachedBundle.RelocationFilePath

	if opts.Name == "" {
		opts.Name = cachedBundle.Definition.Name
	}

	return cachedBundle, nil
}
