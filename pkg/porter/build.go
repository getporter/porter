package porter

import (
	"fmt"
	"os"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/build"
	configadapter "github.com/deislabs/porter/pkg/cnab/config_adapter"
	"github.com/pkg/errors"
)

type BuildProvider interface {
	// BuildInvocationImage using the bundle in the current directory
	BuildInvocationImage() error
}

type BuildOptions struct {
	contextOptions
}

func (p *Porter) Build(opts BuildOptions) error {
	opts.Apply(p.Context)

	err := p.LoadManifest()
	if err != nil {
		return err
	}

	generator := build.NewDockerfileGenerator(p.Config, p.Templates)

	if err := generator.PrepareFilesystem(); err != nil {
		return fmt.Errorf("unable to copy mixins: %s", err)
	}
	if err := generator.GenerateDockerFile(); err != nil {
		return fmt.Errorf("unable to generate Dockerfile: %s", err)
	}
	err = p.Builder.BuildInvocationImage()
	if err != nil {
		return errors.Wrap(err, "unable to build CNAB invocation image")
	}

	return p.buildBundle(p.Config.Manifest.Image, "")
}

func (p *Porter) buildBundle(invocationImage string, digest string) error {
	converter := configadapter.ManifestConverter{Config: p.Config}
	bun := converter.ToBundle()
	return p.writeBundle(bun)
}

func (p Porter) writeBundle(b *bundle.Bundle) error {
	f, err := p.Config.FileSystem.OpenFile(build.LOCAL_BUNDLE, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return errors.Wrapf(err, "error creating %s", build.LOCAL_BUNDLE)
	}
	_, err = b.WriteTo(f)
	return errors.Wrapf(err, "error writing to %s", build.LOCAL_BUNDLE)
}
