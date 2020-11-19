package porter

import (
	"fmt"
	"os"

	"get.porter.sh/porter/pkg/build"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/printer"
	"github.com/Masterminds/semver"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

type BuildProvider interface {
	// BuildInvocationImage using the bundle in the current directory
	BuildInvocationImage(manifest *manifest.Manifest) error
}

type BuildOptions struct {
	contextOptions
	metadataOpts
	NoLint bool
}

func (o BuildOptions) Validate() error {
	if o.Name != "" {
		// TODO: What sort of validation, if any, do we want on the bundle name?
		// Originally I had used the *installation* validation from cnab-go,
		// but that is most likely more restrictive than what we want (no spaces allowed, etc.)
	}

	if o.Version != "" {
		if _, err := semver.NewVersion(o.Version); err != nil {
			return errors.Wrapf(err, "invalid bundle version %q.  Cannot be parsed as semver", o.Version)
		}
	}

	return nil
}

func (p *Porter) Build(opts BuildOptions) error {
	opts.Apply(p.Context)

	if err := opts.Validate(); err != nil {
		return err
	}

	if err := p.generateInternalManifest(opts.metadataOpts); err != nil {
		return errors.Wrap(err, "unable to generate manifest")
	}

	// Publish may invoke this method and the manifest will already be
	// populated.  Only load if still empty.
	if p.Manifest == nil {
		if err := p.LoadManifestFrom(build.LOCAL_MANIFEST); err != nil {
			return err
		}
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
		for _, m := range p.Manifest.Mixins {
			if installedMixin.Name == m.Name {
				usedMixins = append(usedMixins, installedMixin)
			}
		}
	}

	return usedMixins, nil
}

func (p *Porter) buildBundle(invocationImage string, digest string) error {
	imageDigests := map[string]string{invocationImage: digest}

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

func (p Porter) writeBundle(b bundle.Bundle) error {
	f, err := p.Config.FileSystem.OpenFile(build.LOCAL_BUNDLE, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return errors.Wrapf(err, "error creating %s", build.LOCAL_BUNDLE)
	}
	_, err = b.WriteTo(f)
	return errors.Wrapf(err, "error writing to %s", build.LOCAL_BUNDLE)
}
