package porter

import (
	"fmt"
	"os"

	"get.porter.sh/porter/pkg/build"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

type BuildProvider interface {
	// BuildInvocationImage using the bundle in the current directory
	BuildInvocationImage(manifest *manifest.Manifest) error
}

type BuildOptions struct {
	contextOptions
	NoLint bool
}

func (p *Porter) Build(opts BuildOptions) error {
	opts.Apply(p.Context)

	err := p.LoadManifest()
	if err != nil {
		return err
	}

	if !opts.NoLint {
		err = p.preLint()
		if err != nil {
			return err
		}
	}

	generator := build.NewDockerfileGenerator(p.Config, p.Manifest, p.Templates, p.Mixins)

	if err := generator.PrepareFilesystem(); err != nil {
		return fmt.Errorf("unable to copy mixins: %s", err)
	}
	if err := generator.GenerateDockerFile(); err != nil {
		return fmt.Errorf("unable to generate Dockerfile: %s", err)
	}
	if err := p.Builder.BuildInvocationImage(p.Manifest); err != nil {
		return errors.Wrap(err, "unable to build CNAB invocation image")
	}

	return p.buildBundle(p.Manifest.Image, "")
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
