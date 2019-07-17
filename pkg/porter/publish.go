package porter

import (
	"bytes"
	"fmt"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/build"
	portercontext "github.com/deislabs/porter/pkg/context"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
)

// PublishOptions are options that may be specified when publishing a bundle.
// Porter handles defaulting any missing values.
type PublishOptions struct {
	bundleFileOptions
	InsecureRegistry bool
}

// Validate performs validation on the publish options
func (o *PublishOptions) Validate(cxt *portercontext.Context) error {
	err := o.bundleFileOptions.Validate(cxt)
	if err != nil {
		return err
	}

	if o.File == "" {
		return errors.New("could not find porter.yaml in the current directory, make sure you are in the right directory or specify the porter manifest with --file")
	}

	return nil
}

// Publish is a composite function that publishes an invocation image, rewrites the porter manifest
// and then regenerates the bundle.json. Finally it publishes the manifest to an OCI registry.
func (p *Porter) Publish(opts PublishOptions) error {
	if opts.File != "" {
		err := p.LoadManifestFrom(opts.File)
		if err != nil {
			return err
		}
	}

	if p.Config.Manifest.BundleTag == "" {
		return errors.New("porter.yaml must specify a `tag` value for this bundle")
	}

	err := p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	digest, err := p.Registry.PushInvocationImage(p.Config.Manifest.Image)
	if err != nil {
		return errors.Wrap(err, "unable to push CNAB invocation image")
	}

	bun, err := p.rewriteBundleWithInvocationImageDigest(digest)
	if err != nil {
		return err
	}

	return p.Registry.PushBundle(bun, p.Manifest.BundleTag, opts.InsecureRegistry)
}

func (p *Porter) rewriteBundleWithInvocationImageDigest(digest string) (*bundle.Bundle, error) {
	taggedImage, err := p.rewriteImageWithDigest(p.Config.Manifest.Image, digest)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update invocation image reference: %s")
	}

	fmt.Fprintln(p.Out, "\nRewriting CNAB bundle.json...")
	err = p.buildBundle(taggedImage, digest)
	if err != nil {
		return nil, errors.Wrap(err, "unable to rewrite CNAB bundle.json with updated invocation image digest")
	}

	b, err := p.FileSystem.ReadFile(build.LOCAL_BUNDLE)
	bun, err := bundle.ParseReader(bytes.NewBuffer(b))
	if err != nil {
		return nil, errors.Wrap(err, "unable to load CNAB bundle")
	}

	return &bun, nil
}

func (p *Porter) rewriteImageWithDigest(InvocationImage string, digest string) (string, error) {
	ref, err := reference.Parse(InvocationImage)
	if err != nil {
		return "", fmt.Errorf("unable to parse docker image: %s", err)
	}
	named, ok := ref.(reference.Named)
	if !ok {
		return "", fmt.Errorf("had an issue with the docker image")
	}
	return fmt.Sprintf("%s@%s", named.Name(), digest), nil
}
