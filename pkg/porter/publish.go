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
	"github.com/pivotal/image-relocation/pkg/image"
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
		tag = p.Manifest.BundleTag
	}
	if p.Manifest.BundleTag == "" {
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

func (p *Porter) publishFromArchive(opts PublishOptions) error {
	if p.Debug {
		fmt.Fprintf(p.Err, "Extracting bundle from archive %s...\n", opts.ArchiveFile)
	}
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
	imp := packager.NewImporter(source, tmpDir, l)

	err = imp.Import()
	if err != nil {
		return errors.Wrapf(err, "failed to extract bundle from archive %s", opts.ArchiveFile)
	}

	bun, err := l.Load(filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"), "bundle.json"))
	if err != nil {
		return errors.Wrapf(err, "failed to load bundle from archive %s", opts.ArchiveFile)
	}

	if p.Debug {
		fmt.Fprintf(p.Err, "Publishing bundle %s with tag %s...\n", bun.Name, opts.Tag)
	}

	// Update the bundle with new images (name, digest) based on the original,
	// using the provided bundle tag to derive registry and org
	for i, invImg := range bun.InvocationImages {
		err := p.updateBundleWithNewImage(bun, invImg.Image, opts.Tag, i)
		if err != nil {
			return err
		}
	}
	for name, img := range bun.Images {
		err := p.updateBundleWithNewImage(bun, img.Image, opts.Tag, name)
		if err != nil {
			return err
		}
	}

	return p.Registry.PushBundle(bun, opts.Tag, opts.InsecureRegistry)
}

// updateBundleWithNewImage updates a bundle with a new image at the provided index
// constructed using the provided original image and new bundle tag
func (p *Porter) updateBundleWithNewImage(bun *bundle.Bundle, img, tag string, index interface{}) error {
	newImg, err := getNewImageNameFromBundleTag(img, tag)
	if err != nil {
		return err
	}

	digest, err := p.Registry.Copy(img, newImg)
	if err != nil {
		return err
	}

	taggedImage, err := p.rewriteImageWithDigest(newImg, digest)
	if err != nil {
		return errors.Wrapf(err, "unable to update image reference for %s", img)
	}

	// update bundle with new image
	switch v := index.(type) {
	case int: // invocation images is a slice, indexed by an integer
		i := index.(int)
		origImg := bun.InvocationImages[i]
		updatedImg := origImg.DeepCopy()
		updatedImg.Image = taggedImage
		updatedImg.Digest = digest
		bun.InvocationImages[i] = *updatedImg
	case string: // images is a map, indexed by a string
		i := index.(string)
		origImg := bun.Images[i]
		updatedImg := origImg.DeepCopy()
		updatedImg.Image = taggedImage
		updatedImg.Digest = digest
		bun.Images[i] = *updatedImg
	default:
		return fmt.Errorf("unknown image index type: %v", v)
	}

	return nil
}

// getNewImageNameFromBundleTag derives a new image name from the provided original
// using the provided bundleTag to glean registry/org/etc.
func getNewImageNameFromBundleTag(origImg, bundleTag string) (string, error) {
	// Convert strings to structured image.Name objects
	origImgName, err := image.NewName(origImg)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse image %q into domain/path components", origImg)
	}
	bundleTagName, err := image.NewName(bundleTag)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse bundle tag %q into domain/path components", bundleTag)
	}

	// Swap out Host
	newImg := strings.Replace(origImgName.String(), origImgName.Host(), bundleTagName.Host(), -1)

	// Swap out org (via Path)
	origPathParts := strings.Split(origImgName.Path(), "/")
	tagPathParts := strings.Split(bundleTagName.Path(), "/")
	newOrg := tagPathParts[0]
	if len(origPathParts) == 1 {
		// original image has no org, e.g. a library image
		// so just prepend new org
		newImg = strings.Join([]string{newOrg, newImg}, "/")
	} else {
		newImg = strings.Replace(newImg, origPathParts[0], newOrg, -1)
	}

	return newImg, nil
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
