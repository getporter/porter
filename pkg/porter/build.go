package porter

import (
	"context"
	"errors"
	"fmt"
	"os"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/Masterminds/semver/v3"
	"github.com/opencontainers/go-digest"
	"golang.org/x/sync/errgroup"
)

type BuildOptions struct {
	bundleFileOptions
	metadataOpts
	build.BuildImageOptions

	// NoLint indicates if lint should be run before build.
	NoLint bool

	// Driver to use when building the invocation image.
	Driver string

	// Custom is the unparsed list of NAME=VALUE custom inputs set on the command line.
	Customs []string

	// parsedCustoms is the parsed set of custom inputs from Customs.
	parsedCustoms map[string]string
}

const BuildDriverDefault = config.BuildDriverBuildkit

var BuildDriverAllowedValues = []string{config.BuildDriverBuildkit}

func (o *BuildOptions) Validate(p *Porter) error {
	if o.Version != "" {
		v, err := semver.NewVersion(o.Version)
		if err != nil {
			return fmt.Errorf("invalid bundle version: %q is not a valid semantic version", o.Version)
		}
		o.Version = v.String()
	}

	if o.Driver == "" {
		o.Driver = p.GetBuildDriver()
	}
	if !stringSliceContains(BuildDriverAllowedValues, o.Driver) {
		return fmt.Errorf("invalid --driver value %s", o.Driver)
	}

	// Syncing value back to the config and we will always use the config
	// to determine the driver
	// This would be less awkward if we didn't do an automatic build during publish
	p.Data.BuildDriver = o.Driver

	err := o.parseCustomInputs()
	if err != nil {
		return err
	}

	return o.bundleFileOptions.Validate(p.Context)
}

func stringSliceContains(allowedValues []string, value string) bool {
	for _, allowed := range allowedValues {
		if value == allowed {
			return true
		}
	}
	return false
}

func (o *BuildOptions) parseCustomInputs() error {
	p, err := storage.ParseVariableAssignments(o.Customs)
	if err != nil {
		return err
	}

	o.parsedCustoms = p

	return nil
}

func (p *Porter) Build(ctx context.Context, opts BuildOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debugf("Using %s build driver", p.GetBuildDriver())

	// Start with a fresh .cnab directory before building
	err := p.FileSystem.RemoveAll(build.LOCAL_CNAB)
	if err != nil {
		return span.Error(fmt.Errorf("could not cleanup generated .cnab directory before building: %w", err))
	}

	// Generate Porter's canonical version of the user-provided manifest
	if err := p.generateInternalManifest(ctx, opts); err != nil {
		return fmt.Errorf("unable to generate manifest: %w", err)
	}

	m, err := manifest.LoadManifestFrom(ctx, p.Config, build.LOCAL_MANIFEST)
	if err != nil {
		return err
	}

	// Capture the path to the original, user-provided manifest.
	// This value will be referenced elsewhere, for instance by
	// the digest logic (to dictate auto-rebuild)
	m.ManifestPath = opts.File

	if !opts.NoLint {
		if err := p.preLint(ctx, opts.File); err != nil {
			return err
		}
	}

	// Build bundle so that resulting bundle.json is available for inclusion
	// into the invocation image.
	// Note: the content digest field on the invocation image section of the
	// bundle.json will *not* be correct until the image is actually pushed
	// to a registry.  The bundle.json will need to be updated after publishing
	// and provided just-in-time during bundle execution.
	if err := p.buildBundle(ctx, m, ""); err != nil {
		return span.Error(fmt.Errorf("unable to build bundle: %w", err))
	}

	generator := build.NewDockerfileGenerator(p.Config, m, p.Templates, p.Mixins)

	if err := generator.PrepareFilesystem(); err != nil {
		return span.Error(fmt.Errorf("unable to copy run script, runtimes or mixins: %s", err))
	}
	if err := generator.GenerateDockerFile(ctx); err != nil {
		return span.Error(fmt.Errorf("unable to generate Dockerfile: %s", err))
	}

	builder := p.GetBuilder(ctx)

	err = builder.BuildInvocationImage(ctx, m, opts.BuildImageOptions)
	if err != nil {
		return span.Error(fmt.Errorf("unable to build CNAB invocation image: %w", err))
	}

	return nil
}

func (p *Porter) preLint(ctx context.Context, file string) error {
	lintOpts := LintOptions{
		PrintOptions: printer.PrintOptions{},
		File:         file,
	}
	lintOpts.RawFormat = string(printer.FormatPlaintext)
	err := lintOpts.Validate(p.Context)
	if err != nil {
		return err
	}

	results, err := p.Lint(ctx, lintOpts)
	if err != nil {
		return err
	}

	if len(results) > 0 {
		fmt.Fprintln(p.Out, results.String())
	}

	if results.HasError() {
		// An error was found during linting, stop and let the user correct it
		return errors.New("lint errors were detected. Rerun with --no-lint ignore the errors")
	}

	return nil
}

func (p *Porter) getUsedMixins(ctx context.Context, m *manifest.Manifest) ([]mixin.Metadata, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	g := new(errgroup.Group)
	results := make(chan mixin.Metadata, len(m.Mixins))
	for _, m := range m.Mixins {
		m := m
		g.Go(func() error {
			result, err := p.Mixins.GetMetadata(ctx, m.Name)
			if err != nil {
				return err
			}

			mixinMetadata := result.(*mixin.Metadata)
			results <- *mixinMetadata
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	usedMixins := make([]mixin.Metadata, len(m.Mixins))
	for i := 0; i < len(usedMixins); i++ {
		result := <-results
		usedMixins[i] = result
	}

	return usedMixins, nil
}

func (p *Porter) buildBundle(ctx context.Context, m *manifest.Manifest, digest digest.Digest) error {
	imageDigests := map[string]string{m.Image: digest.String()}

	mixins, err := p.getUsedMixins(ctx, m)
	if err != nil {
		return err
	}

	converter := configadapter.NewManifestConverter(p.Config, m, imageDigests, mixins)
	bun, err := converter.ToBundle(ctx)
	if err != nil {
		return err
	}

	return p.writeBundle(bun)
}

func (p Porter) writeBundle(b cnab.ExtendedBundle) error {
	f, err := p.Config.FileSystem.OpenFile(build.LOCAL_BUNDLE, os.O_RDWR|os.O_CREATE|os.O_TRUNC, pkg.FileModeWritable)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", build.LOCAL_BUNDLE, err)
	}
	defer f.Close()
	_, err = b.WriteTo(f)
	if err != nil {
		return fmt.Errorf("error writing to %s: %w", build.LOCAL_BUNDLE, err)
	}

	return nil
}
