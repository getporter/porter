package porter

import (
	"fmt"
	"os"
	"regexp"

	"get.porter.sh/porter/pkg/build"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

type BuildProvider interface {
	// BuildInvocationImage using the bundle in the build context directory
	BuildInvocationImage(manifest *manifest.Manifest) error

	// TagInvocationImage using the origTag and newTag values supplied
	TagInvocationImage(origTag, newTag string) error
}

type BuildOptions struct {
	bundleFileOptions
	contextOptions
	metadataOpts
	NoLint bool
}

// semVerRegex is a regex for ensuring bundle versions adhere to
// semantic versioning per https://semver.org/#is-v123-a-semantic-version
// Regex adapted from github.com/Masterminds/semver
const semVerRegex string = `([0-9]+)(\.[0-9]+)?(\.[0-9]+)?` +
	`(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
	`(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?`

func (o *BuildOptions) Validate(cxt *context.Context) error {
	if o.Version != "" {
		versionRegex := regexp.MustCompile("^" + semVerRegex + "$")
		if m := versionRegex.FindStringSubmatch(o.Version); m == nil {
			return fmt.Errorf("invalid bundle version: %q is not a valid semantic version", o.Version)
		}
	}

	return o.bundleFileOptions.Validate(cxt)
}

func (p *Porter) Build(opts BuildOptions) error {
	opts.Apply(p.Context)

	if err := opts.Validate(p.Context); err != nil {
		return err
	}

	if err := p.generateInternalManifest(opts); err != nil {
		return errors.Wrap(err, "unable to generate manifest")
	}

	if err := p.LoadManifestFrom(build.LOCAL_MANIFEST); err != nil {
		return err
	}

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
	if err := p.buildBundle(p.Manifest.Image, "", opts.File); err != nil {
		return errors.Wrap(err, "unable to build bundle")
	}

	generator := build.NewDockerfileGenerator(p.Config, p.Manifest, opts.File, p.Templates, p.Mixins)

	if err := generator.PrepareFilesystem(); err != nil {
		return fmt.Errorf("unable to copy run script, runtimes or mixins: %s", err)
	}
	if err := generator.GenerateDockerFile(); err != nil {
		return fmt.Errorf("unable to generate Dockerfile: %s", err)
	}

	return errors.Wrap(p.Builder.BuildInvocationImage(p.Manifest), "unable to build CNAB invocation image")
}

func (p *Porter) preLint() error {
	lintOpts := LintOptions{}
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
		if p.Manifest == nil {
			return usedMixins, nil
		}
		for _, m := range p.Manifest.Mixins {
			if installedMixin.Name == m.Name {
				usedMixins = append(usedMixins, installedMixin)
			}
		}
	}

	return usedMixins, nil
}

func (p *Porter) buildBundle(invocationImage string, digest string, manifestPath string) error {
	fmt.Fprintf(p.Out, "wd = %s\n", p.Getwd())
	imageDigests := map[string]string{invocationImage: digest}

	mixins, err := p.getUsedMixins()

	if err != nil {
		return err
	}

	converter := configadapter.NewManifestConverter(p.Context, p.Manifest, manifestPath, imageDigests, mixins)
	bun, err := converter.ToBundle()
	if err != nil {
		return err
	}

	return p.writeBundle(bun)
}

func (p Porter) writeBundle(b bundle.Bundle) error {
	f, err := p.Config.FileSystem.OpenFile(build.LOCAL_BUNDLE, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return errors.Wrapf(err, "error creating %s", build.LOCAL_BUNDLE)
	}
	_, err = b.WriteTo(f)
	return errors.Wrapf(err, "error writing to %s", build.LOCAL_BUNDLE)
}
