package porter

import (
	"context"
	"errors"
	"fmt"

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
	"go.mongodb.org/mongo-driver/bson"
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

	// parameters that are intended for dependencies
	// This is legacy support for v1 of dependencies where you could pass a parameter to a dependency directly using special formatting
	// Example: --param mysql#username=admin
	// This is not used anymore in dependencies v2
	depParams map[string]string

	// A cache of the final resolved set of parameters that are passed to the bundle
	// Do not use directly, use GetParameters instead.
	finalParams map[string]interface{}
}

func NewBundleExecutionOptions() *BundleExecutionOptions {
	return &BundleExecutionOptions{
		BundleReferenceOptions: &BundleReferenceOptions{},
	}
}

func (o *BundleExecutionOptions) GetOptions() *BundleExecutionOptions {
	return o
}

// GetParameters returns the final resolved set of a parameters to pass to the bundle.
// You must have already called Porter.applyActionOptionsToInstallation to populate this value as
// this just returns the cached set of parameters
func (o *BundleExecutionOptions) GetParameters() map[string]interface{} {
	if o.finalParams == nil {
		panic("BundleExecutionOptions.GetParameters was called before the final set of parameters were resolved with Porter.applyActionOptionsToInstallation")
	}
	return o.finalParams
}

func (o *BundleExecutionOptions) Validate(ctx context.Context, args []string, p *Porter) error {
	if err := o.BundleReferenceOptions.Validate(ctx, args, p); err != nil {
		return err
	}

	o.defaultDriver(p)
	if err := o.validateDriver(p.Context); err != nil {
		return err
	}

	return nil
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

	// DO NOT ACCESS DIRECTLY, use GetBundleReference to retrieve and cache the value
	bundleRef *cnab.BundleReference
}

// GetBundleReference resolves the bundle reference if needed and caches the result so that this is safe to call multiple times in a row.
func (o *BundleReferenceOptions) GetBundleReference(ctx context.Context, p *Porter) (cnab.BundleReference, error) {
	if o.bundleRef == nil {
		ref, err := p.resolveBundleReference(ctx, o)
		if err != nil {
			return cnab.BundleReference{}, err
		}

		o.bundleRef = &ref
	}

	return *o.bundleRef, nil
}

// UnsetBundleReference clears the cached bundle reference so that it may be re-resolved the next time GetBundleReference is called.
func (o *BundleReferenceOptions) UnsetBundleReference() {
	o.bundleRef = nil
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

// resolveBundleReference uses the bundle options from the CLI flags to determine which bundle is being referenced.
// Takes into account the --reference, --file and --cnab-file flags, and also uses the NAME argument and looks up the bundle definition from the installation.
// Do not call this directly. Call BundleReferenceOptions.GetBundleReference() instead so that it's safe to call multiple times in a row and returns a cached results after being resolved.
func (p *Porter) resolveBundleReference(ctx context.Context, opts *BundleReferenceOptions) (cnab.BundleReference, error) {
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
		buildOpts := BuildOptions{
			BundleDefinitionOptions: opts.BundleDefinitionOptions,
			InsecureRegistry:        opts.InsecureRegistry,
		}
		localBundle, err := p.ensureLocalBundleIsUpToDate(ctx, buildOpts)
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

	return bundleRef, nil
}

// BuildActionArgs converts an instance of user-provided action options into prepared arguments
// that can be used to execute the action.
func (p *Porter) BuildActionArgs(ctx context.Context, inst storage.Installation, action BundleAction) (cnabprovider.ActionArguments, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	opts := action.GetOptions()
	bundleRef, err := opts.GetBundleReference(ctx, p)
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	if opts.RelocationMapping != "" {
		err := encoding.UnmarshalFile(p.FileSystem, opts.RelocationMapping, &bundleRef.RelocationMap)
		if err != nil {
			return cnabprovider.ActionArguments{}, span.Errorf("could not parse the relocation mapping file at %s: %w", opts.RelocationMapping, err)
		}
	}

	run, err := p.createRun(ctx, bundleRef, inst, action.GetAction(), opts.GetParameters())
	if err != nil {
		return cnabprovider.ActionArguments{}, err
	}

	args := cnabprovider.ActionArguments{
		Run:                   run,
		Installation:          inst,
		BundleReference:       bundleRef,
		Params:                opts.GetParameters(),
		Driver:                opts.Driver,
		AllowDockerHostAccess: opts.AllowDockerHostAccess,
		PersistLogs:           !opts.NoLogs,
	}

	return args, nil
}

// createRun generates a Run record instructing porter exactly how to run the bundle
// and includes audit/status fields as well.
func (p *Porter) createRun(ctx context.Context, bundleRef cnab.BundleReference, inst storage.Installation, action string, params map[string]interface{}) (storage.Run, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// Create a record for the run we are about to execute
	var currentRun = inst.NewRun(action, bundleRef.Definition)
	currentRun.Bundle = bundleRef.Definition.Bundle
	currentRun.BundleReference = bundleRef.Reference.String()
	currentRun.BundleDigest = bundleRef.Digest.String()

	var err error
	cleanParams, err := p.Sanitizer.CleanRawParameters(ctx, params, bundleRef.Definition, currentRun.ID)
	if err != nil {
		return storage.Run{}, span.Error(err)
	}
	currentRun.Parameters.Parameters = cleanParams

	// TODO: Do not save secrets when the run isn't recorded
	currentRun.ParameterOverrides = storage.LinkSensitiveParametersToSecrets(currentRun.ParameterOverrides, bundleRef.Definition, currentRun.ID)
	currentRun.ParameterSets = inst.ParameterSets

	// Persist an audit record of the credential sets used to determine the final
	// credentials injected into the bundle.
	//
	// These should remain in the order specified on the installation, and not
	// sorted, so that the last specified set overrides the one before it when a
	// value is specified in more than one set.
	currentRun.CredentialSets = inst.CredentialSets

	// Combine the credential sets above into a single credential set we can resolve just-in-time (JIT) before running the bundle.
	finalCreds := make(map[string]secrets.Strategy, len(currentRun.Bundle.Credentials))
	for _, csName := range currentRun.CredentialSets {
		var cs storage.CredentialSet
		// Try to get the creds in the local namespace first, fallback to the global creds
		query := storage.FindOptions{
			Sort: []string{"-namespace"},
			Filter: bson.M{
				"name": csName,
				"$or": []bson.M{
					{"namespace": ""},
					{"namespace": currentRun.Namespace},
				},
			},
		}
		store := p.Credentials.GetDataStore()
		err := store.FindOne(ctx, storage.CollectionCredentials, query, &cs)
		if err != nil {
			return storage.Run{}, span.Errorf("could not find credential set named %s in the %s namespace or global namespace: %w", csName, inst.Namespace, err)
		}

		for _, cred := range cs.Credentials {
			credDef, ok := currentRun.Bundle.Credentials[cred.Name]
			if !ok || !credDef.AppliesTo(currentRun.Action) {
				// ignore extra credential mappings in the set that are not defined by the bundle or used by the current action
				// it's okay to over specify so that people can reuse sets better
				continue
			}

			// If a credential is mapped in multiple credential sets, the strategy associated with the last set specified "wins"
			finalCreds[cred.Name] = cred
		}
	}

	if len(finalCreds) > 0 {
		// Store the composite credential set on the run, so that the runtime can later resolve them in a single step
		currentRun.Credentials = storage.NewInternalCredentialSet()
		for _, cred := range finalCreds {
			currentRun.Credentials.Credentials = append(currentRun.Credentials.Credentials, cred)
		}
	}

	return currentRun, nil
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
