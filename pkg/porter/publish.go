package porter

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/loader"
	"github.com/deislabs/cnab-go/packager"
	"github.com/docker/cnab-to-oci/relocation"
	"github.com/docker/distribution/reference"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
	"github.com/pivotal/image-relocation/pkg/registry/ggcr"
	"github.com/pkg/errors"

	"get.porter.sh/porter/pkg/build"
	portercontext "get.porter.sh/porter/pkg/context"
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

	rm, err := p.Registry.PushBundle(bun, tag, opts.InsecureRegistry)
	if err != nil {
		return err
	}

	// Perhaps we have a cached version of a bundle with the same tag, previously pulled
	// If so, replace it, as it is most likely out-of-date per this publish
	return p.refreshCachedBundle(bun, tag, rm)
}

// publishFromArchive (re-)publishes a bundle, provided by the archive file, using the provided tag.
//
// After the bundle is extracted from the archive, we iterate through all of the images (invocation
// and application) listed in the bundle, grab their digests by parsing the extracted
// OCI Layout, rename each based on the registry/org values derived from the provided tag
// and then push each updated image with the original digests
//
// Finally, we generate a new bundle from the old, with all image names and digests updated, based
// on the newly copied images, and then push this new bundle using the provided tag.
// (Currently we use the docker/cnab-to-oci library for this logic.)
//
// In the generation of a new bundle, we therefore don't preserve content digests and can't maintain
// signature verification throughout the process.  Once we wish to preserve content digest and such verification,
// this approach will need to be refactored, via preserving the original bundle and employing
// a relocation mapping approach to associate the bundle's (old) images with the newly copied images.
func (p *Porter) publishFromArchive(opts PublishOptions) error {
	source, err := filepath.Abs(opts.ArchiveFile)
	if err != nil {
		return errors.Wrapf(err, "could not determine absolute path to archive file %s", opts.ArchiveFile)
	}

	tmpDir, err := p.FileSystem.TempDir("", "porter")
	if err != nil {
		return errors.Wrap(err, "error creating temp directory for archive extraction")
	}
	defer p.FileSystem.RemoveAll(tmpDir)
	extractedDir := filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"))

	bun, err := p.extractBundle(tmpDir, source)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "Beginning bundle publish to %s. This may take some time.\n", opts.Tag)

	// Use the ggcr client to read the extracted OCI Layout
	client := ggcr.NewRegistryClient()
	layout, err := client.ReadLayout(filepath.Join(extractedDir, "artifacts/layout"))
	if err != nil {
		return errors.Wrapf(err, "failed to parse OCI Layout from archive %s", opts.ArchiveFile)
	}

	// Push updated images (renamed based on provided bundle tag) with same digests
	// then update the bundle with new values (image name, digest)
	for i, invImg := range bun.InvocationImages {
		newImgName, err := getNewImageNameFromBundleTag(invImg.Image, opts.Tag)
		if err != nil {
			return err
		}

		digest, err := pushUpdatedImage(layout, invImg.Image, newImgName)
		if err != nil {
			return err
		}

		err = p.updateBundleWithNewImage(bun, newImgName, digest, i)
		if err != nil {
			return err
		}
	}
	for name, img := range bun.Images {
		newImgName, err := getNewImageNameFromBundleTag(img.Image, opts.Tag)
		if err != nil {
			return err
		}

		digest, err := pushUpdatedImage(layout, img.Image, newImgName)
		if err != nil {
			return err
		}

		err = p.updateBundleWithNewImage(bun, newImgName, digest, name)
		if err != nil {
			return err
		}
	}

	rm, err := p.Registry.PushBundle(bun, opts.Tag, opts.InsecureRegistry)
	if err != nil {
		return err
	}

	// Perhaps we have a cached version of a bundle with the same tag, previously pulled
	// If so, replace it, as it is most likely out-of-date per this publish
	return p.refreshCachedBundle(bun, opts.Tag, rm)
}

// extractBundle extracts a bundle using the provided opts and returnsthe extracted bundle
func (p *Porter) extractBundle(tmpDir, source string) (*bundle.Bundle, error) {
	if p.Debug {
		fmt.Fprintf(p.Err, "Extracting bundle from archive %s...\n", source)
	}

	l := loader.NewLoader()
	imp := packager.NewImporter(source, tmpDir, l)
	err := imp.Import()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to extract bundle from archive %s", source)
	}

	bun, err := l.Load(filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"), "bundle.json"))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load bundle from archive %s", source)
	}

	return bun, nil
}

// pushUpdatedImage uses the provided layout to find the provided origImg,
// gathers the pre-existing digest and then pushes this digest using the newImgName
func pushUpdatedImage(layout registry.Layout, origImg string, newImgName image.Name) (image.Digest, error) {
	origImgName, err := image.NewName(origImg)
	if err != nil {
		return image.EmptyDigest, errors.Wrapf(err, "unable to parse image %q into domain/path components", origImg)
	}

	digest, err := layout.Find(origImgName)
	if err != nil {
		return image.EmptyDigest, errors.Wrapf(err, "unable to find image %s in archived OCI Layout", origImgName.String())
	}

	err = layout.Push(digest, newImgName)
	if err != nil {
		return image.EmptyDigest, errors.Wrapf(err, "unable to push image %s", newImgName.String())
	}

	return digest, nil
}

// updateBundleWithNewImage updates a bundle with a new image (with digest) at the provided index
func (p *Porter) updateBundleWithNewImage(bun *bundle.Bundle, newImg image.Name, digest image.Digest, index interface{}) error {
	taggedImage, err := p.rewriteImageWithDigest(newImg.String(), digest.String())
	if err != nil {
		return errors.Wrapf(err, "unable to update image reference for %s", newImg.String())
	}

	// update bundle with new image
	switch v := index.(type) {
	case int: // invocation images is a slice, indexed by an integer
		i := index.(int)
		origImg := bun.InvocationImages[i]
		updatedImg := origImg.DeepCopy()
		updatedImg.Image = taggedImage
		updatedImg.Digest = digest.String()
		bun.InvocationImages[i] = *updatedImg
	case string: // images is a map, indexed by a string
		i := index.(string)
		origImg := bun.Images[i]
		updatedImg := origImg.DeepCopy()
		updatedImg.Image = taggedImage
		updatedImg.Digest = digest.String()
		bun.Images[i] = *updatedImg
	default:
		return fmt.Errorf("unknown image index type: %v", v)
	}

	return nil
}

// getNewImageNameFromBundleTag derives a new image.Name object from the provided original
// image (string) using the provided bundleTag to glean registry/org/etc.
func getNewImageNameFromBundleTag(origImg, bundleTag string) (image.Name, error) {
	origImgName, err := image.NewName(origImg)
	if err != nil {
		return image.EmptyName, errors.Wrapf(err, "unable to parse image %q into domain/path components", origImg)
	}

	bundleTagName, err := image.NewName(bundleTag)
	if err != nil {
		return image.EmptyName, errors.Wrapf(err, "unable to parse bundle tag %q into domain/path components", bundleTag)
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

	newImgName, err := image.NewName(newImg)
	if err != nil {
		return image.EmptyName, errors.Wrapf(err, "unable to parse image %q into domain/path components", newImg)
	}

	return newImgName, nil
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

// refreshCachedBundle will store a bundle anew, if a bundle with the same tag is found in the cache
func (p *Porter) refreshCachedBundle(bun *bundle.Bundle, tag string, rm relocation.ImageRelocationMap) error {
	if _, _, found, _ := p.Cache.FindBundle(tag); found {
		_, _, err := p.Cache.StoreBundle(tag, bun, rm)
		if err != nil {
			fmt.Fprintf(p.Err, "warning: unable to update cache for bundle %s: %s\n", tag, err)
		}
	}
	return nil
}
