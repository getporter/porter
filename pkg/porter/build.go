package porter

import (
	"fmt"
	"os"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/printer"
	"github.com/Masterminds/semver/v3"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

type BuildOptions struct {
	bundleFileOptions
	contextOptions
	metadataOpts

	// NoLint indicates if lint should be run before build.
	NoLint bool

	// Driver to use when building the invocation image.
	Driver string
}

const BuildDriverDefault = config.BuildDriverDocker

var BuildDriverAllowedValues = []string{config.BuildDriverDocker, config.BuildDriverBuildkit}

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
		return errors.Errorf("invalid --driver value %s", o.Driver)
	}

	// Syncing value back to the config and we will always use the config
	// to determine the driver
	// This would be less awkward if we didn't do an automatic build during publish
	p.Data.BuildDriver = o.Driver

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

func (p *Porter) Build(opts BuildOptions) error {
	opts.Apply(p.Context)

	if p.Debug {
		fmt.Fprintf(p.Err, "Using %s build driver\n", p.GetBuildDriver())
	}

	// Start with a fresh .cnab directory before building
	err := p.FileSystem.RemoveAll(build.LOCAL_CNAB)
	if err != nil {
		return errors.Wrap(err, "could not cleanup generated .cnab directory before building")
	}

	// Generate Porter's canonical version of the user-provided manifest
	if err := p.generateInternalManifest(opts); err != nil {
		return errors.Wrap(err, "unable to generate manifest")
	}

	if err := p.LoadManifestFrom(build.LOCAL_MANIFEST); err != nil {
		return err
	}

	// Capture the path to the original, user-provided manifest.
	// This value will be referenced elsewhere, for instance by
	// the digest logic (to dictate auto-rebuild)
	p.Manifest.ManifestPath = opts.File

	if !opts.NoLint {
		if err := p.preLint(); err != nil {
			return err
		}
	}

	// Build bundle so that resulting bundle.json is available for inclusion
	// into the invocation image.
	// Note: the content digest field on the invocation image section of the
	// bundle.json will *not* be correct until the image is actually pushed
	// to a registry.  The bundle.json will need to be updated after publishing
	// and provided just-in-time during bundle execution.
	if err := p.buildBundle(p.Manifest.Image, ""); err != nil {
		return errors.Wrap(err, "unable to build bundle")
	}

	generator := build.NewDockerfileGenerator(p.Config, p.Manifest, p.Templates, p.Mixins)

	if err := generator.PrepareFilesystem(); err != nil {
		return fmt.Errorf("unable to copy run script, runtimes or mixins: %s", err)
	}
	if err := generator.GenerateDockerFile(); err != nil {
		return fmt.Errorf("unable to generate Dockerfile: %s", err)
	}

	builder := p.GetBuilder()
	return errors.Wrap(builder.BuildInvocationImage(p.Manifest), "unable to build CNAB invocation image")
}

func (p *Porter) preLint() error {
	lintOpts := LintOptions{
		contextOptions: NewContextOptions(p.Context),
		PrintOptions:   printer.PrintOptions{},
	}
	lintOpts.RawFormat = string(printer.FormatPlaintext)
	err := lintOpts.Validate(p.Context)
	if err != nil {
		return err
	}

	results, err := p.Lint(lintOpts)
	if err != nil {
		return err
	}

	if len(results) > 0 {
		fmt.Fprintln(p.Out, results.String())
	}

	if results.HasError() {
		// An error was found during linting, stop and let the user correct it
		return errors.New("Lint errors were detected. Rerun with --no-lint ignore the errors.")
	}

	return nil
}

func (p *Porter) getUsedMixins() ([]mixin.Metadata, error) {
	installedMixins, err := p.ListMixins()

	if err != nil {
		return nil, errors.Wrapf(err, "error while listing mixins")
	}

	var usedMixins []mixin.Metadata
	for _, installedMixin := range installedMixins {
		for _, m := range p.Manifest.Mixins {
			if installedMixin.Name == m.Name {
				usedMixins = append(usedMixins, installedMixin)
			}
		}
	}

	return usedMixins, nil
}

func (p *Porter) buildBundle(invocationImage string, digest digest.Digest) error {
	imageDigests := map[string]string{invocationImage: digest.String()}

	mixins, err := p.getUsedMixins()

	if err != nil {
		return err
	}

	converter := configadapter.NewManifestConverter(p.Context, p.Manifest, imageDigests, mixins)
	bun, err := converter.ToBundle()
	if err != nil {
		return err
	}

	return p.writeBundle(bun)
}

func (p Porter) writeBundle(b cnab.ExtendedBundle) error {
	f, err := p.Config.FileSystem.OpenFile(build.LOCAL_BUNDLE, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		return errors.Wrapf(err, "error creating %s", build.LOCAL_BUNDLE)
	}
	_, err = b.WriteTo(f)
	return errors.Wrapf(err, "error writing to %s", build.LOCAL_BUNDLE)
}
