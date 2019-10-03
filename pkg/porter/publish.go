package porter

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/loader"
	"github.com/deislabs/cnab-go/packager"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/build"
	portercontext "github.com/deislabs/porter/pkg/context"
)

// PublishOptions are options that may be specified when publishing a bundle.
// Porter handles defaulting any missing values.
type PublishOptions struct {
	BundlePullOptions
	bundleFileOptions
	InsecureRegistry bool
	ArchiveFile      string
}

// Validate performs validation on the publish options
func (o *PublishOptions) Validate(cxt *portercontext.Context) error {
	if o.ArchiveFile != "" {
		// Verify the archive file can be accessed
		path, err := filepath.Abs(o.ArchiveFile)
		if err != nil {
			return errors.Wrapf(err, "unable to determine absolute path for --archive %s", o.ArchiveFile)
		}
		if _, err := cxt.FileSystem.Stat(path); err != nil {
			return errors.Wrapf(err, "unable to access --archive %s", o.ArchiveFile)
		}

		if o.Tag == "" {
			return errors.New("must provide a value for --tag of the form REGISTRY/bundle:tag")
		}
	} else {
		// Proceed with publishing from current directory
		err := o.bundleFileOptions.Validate(cxt)
		if err != nil {
			return err
		}

		if o.File == "" {
			return errors.New("could not find porter.yaml in the current directory, make sure you are in the right directory or specify the porter manifest with --file")
		}
	}

	if o.Tag != "" {
		return o.validateTag()
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

	if opts.ArchiveFile == "" {
		return p.publishFromFile(opts)
	} else {
		return p.publishFromArchive(opts)
	}
}

func (p *Porter) publishFromFile(opts PublishOptions) error {
	tag := opts.Tag
	if tag == "" {
		tag = p.Config.Manifest.BundleTag
	}
	if p.Config.Manifest.BundleTag == "" {
		return errors.New("porter.yaml must specify a `tag` value for this bundle")
	}

	err := p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	digest, err := p.Registry.PushInvocationImage(p.Manifest.Image)
	if err != nil {
		return errors.Wrap(err, "unable to push CNAB invocation image")
	}

	bun, err := p.rewriteBundleWithInvocationImageDigest(digest)
	if err != nil {
		return err
	}

	return p.Registry.PushBundle(bun, tag, opts.InsecureRegistry)
}

// TODO: tests!
func (p *Porter) publishFromArchive(opts PublishOptions) error {
	fmt.Fprintf(p.Out, "Extracting bundle from archive %s...\n", opts.ArchiveFile)
	source, err := filepath.Abs(opts.ArchiveFile)
	if err != nil {
		return errors.Wrapf(err, "could not determine absolute path to archive file %s", opts.ArchiveFile)
	}

	tmpDir, err := p.FileSystem.TempDir("", "porter")
	if err != nil {
		return errors.Wrap(err, "error creating temp directory for archive extraction")
	}
	defer p.FileSystem.RemoveAll(tmpDir)

	l := loader.NewLoader()
	// as of writing, NewImporter always returns nil as error
	imp, _ := packager.NewImporter(source, tmpDir, l, false)

	err = imp.Import()
	if err != nil {
		return errors.Wrapf(err, "failed to extract bundle from archive %s", opts.ArchiveFile)
	}

	// TODO: do we want to import bundle into Porter's cache?

	bun, err := l.Load(filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"), "bundle.json"))
	if err != nil {
		return errors.Wrapf(err, "failed to load bundle from archive %s", opts.ArchiveFile)
	}

	fmt.Fprintf(p.Out, "Publishing bundle %s with tag %s...\n", bun.Name, opts.Tag)
	// TODO: support overriding with new/different registry, etc.?  (image relocation effort?)
	invocationImage := bun.InvocationImages[0].Image
	digest, err := p.Registry.PushInvocationImage(invocationImage)
	if err != nil {
		return errors.Wrap(err, "unable to push CNAB invocation image")
	}

	// update bundle with digest
	taggedImage, err := p.rewriteImageWithDigest(invocationImage, digest)
	if err != nil {
		return errors.Wrap(err, "unable to update invocation image reference")
	}
	bun.InvocationImages[0].Digest = taggedImage

	return p.Registry.PushBundle(bun, opts.Tag, opts.InsecureRegistry)
}

func (p *Porter) rewriteBundleWithInvocationImageDigest(digest string) (*bundle.Bundle, error) {
	taggedImage, err := p.rewriteImageWithDigest(p.Manifest.Image, digest)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update invocation image reference")
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
