package porter

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	portercontext "get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/cnabio/cnab-go/packager"
	"github.com/opencontainers/go-digest"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
	"github.com/pivotal/image-relocation/pkg/registry/ggcr"
	"github.com/pkg/errors"
)

// PublishOptions are options that may be specified when publishing a bundle.
// Porter handles defaulting any missing values.
type PublishOptions struct {
	BundlePullOptions
	bundleFileOptions
	Tag         string
	Registry    string
	ArchiveFile string
}

// Validate performs validation on the publish options
func (o *PublishOptions) Validate(cxt *portercontext.Context) error {
	if o.ArchiveFile != "" {
		// Verify the archive file can be accessed
		if _, err := cxt.FileSystem.Stat(o.ArchiveFile); err != nil {
			return errors.Wrapf(err, "unable to access --archive %s", o.ArchiveFile)
		}

		if o.Reference == "" {
			return errors.New("must provide a value for --reference of the form REGISTRY/bundle:tag")
		}
	} else {
		// Proceed with publishing from the resolved build context directory
		err := o.bundleFileOptions.Validate(cxt)
		if err != nil {
			return err
		}

		if o.File == "" {
			return errors.New("could not find porter.yaml in the current directory, make sure you are in the right directory or specify the porter manifest with --file")
		}
	}

	if o.Reference != "" {
		return o.BundlePullOptions.Validate()
	}

	if o.Tag != "" {
		return o.validateTag()
	}

	return nil
}

// validateTag checks to make sure the supplied tag is of the expected form.
// A previous iteration of this flag was used to designate an entire bundle
// reference.  If we detect this attempted use, we return an error and
// explanation
func (o *PublishOptions) validateTag() error {
	if strings.Contains(o.Tag, ":") || strings.Contains(o.Tag, "@") {
		return errors.New("the --tag flag has been updated to designate just the Docker tag portion of the bundle reference; use --reference for the full bundle reference instead")
	}
	return nil
}

// Publish is a composite function that publishes an invocation image, rewrites the porter manifest
// and then regenerates the bundle.json. Finally it publishes the manifest to an OCI registry.
func (p *Porter) Publish(opts PublishOptions) error {
	if opts.File != "" {
		if err := p.LoadManifestFrom(opts.File); err != nil {
			return err
		}
	}

	if opts.ArchiveFile == "" {
		return p.publishFromFile(opts)
	}
	return p.publishFromArchive(opts)
}

func (p *Porter) publishFromFile(opts PublishOptions) error {
	_, err := p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	// If the manifest file is the default/user-supplied manifest,
	// hot-swap in Porter's canonical translation (if exists) from
	// the .cnab/app directory, as there may be dynamic overrides for
	// the name and version fields to inform invocation image naming.
	canonicalManifest := filepath.Join(opts.Dir, build.LOCAL_MANIFEST)
	canonicalExists, err := p.FileSystem.Exists(canonicalManifest)
	if err != nil {
		return err
	}
	if canonicalExists {
		err := p.LoadManifestFrom(canonicalManifest)
		if err != nil {
			return err
		}
		// We still want the user-provided manifest path to be tracked,
		// not Porter's canonical manifest path, for digest matching/auto-rebuilds
		p.Manifest.ManifestPath = opts.File
	}

	// Capture original invocation image name as it may be updated below
	origInvImg := p.Manifest.Image

	// Check for tag and registry overrides optionally supplied on publish
	if opts.Tag != "" {
		p.Manifest.DockerTag = opts.Tag
	}
	if opts.Registry != "" {
		p.Manifest.Registry = opts.Registry
	}
	// If either are non-empty, null out the reference on the manifest, as
	// it needs to be rebuilt with new values
	if opts.Tag != "" || opts.Registry != "" {
		p.Manifest.Reference = ""
	}

	// Update invocation image and reference with opts.Reference, which may be
	// empty, which is fine - we still may need to pick up tag and/or registry
	// overrides
	if err := p.Manifest.SetInvocationImageAndReference(opts.Reference); err != nil {
		return errors.Wrap(err, "unable to set invocation image name and reference")
	}

	if origInvImg != p.Manifest.Image {
		// Tag it so that it will be known/found by Docker for publishing
		builder := p.GetBuilder()
		if err := builder.TagInvocationImage(origInvImg, p.Manifest.Image); err != nil {
			return err
		}
	}

	if p.Manifest.Reference == "" {
		return errors.New("porter.yaml is missing registry or reference values needed for publishing")
	}

	var bundleRef cnab.BundleReference
	bundleRef.Reference, err = cnab.ParseOCIReference(p.Manifest.Reference)
	if err != nil {
		return errors.Wrapf(err, "invalid reference %s", p.Manifest.Reference)
	}

	bundleRef.Digest, err = p.Registry.PushInvocationImage(p.Manifest.Image)
	if err != nil {
		return errors.Wrapf(err, "unable to push CNAB invocation image %q", p.Manifest.Image)
	}

	bundleRef.Definition, err = p.rewriteBundleWithInvocationImageDigest(bundleRef.Digest, opts.File)
	if err != nil {
		return err
	}

	bundleRef, err = p.Registry.PushBundle(bundleRef, opts.InsecureRegistry)
	if err != nil {
		return err
	}

	// Perhaps we have a cached version of a bundle with the same reference, previously pulled
	// If so, replace it, as it is most likely out-of-date per this publish
	return p.refreshCachedBundle(bundleRef)
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
	source := p.FileSystem.Abs(opts.ArchiveFile)
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

	fmt.Fprintf(p.Out, "Beginning bundle publish to %s. This may take some time.\n", opts.Reference)

	// Use the ggcr client to read the extracted OCI Layout
	client := ggcr.NewRegistryClient()
	layout, err := client.ReadLayout(filepath.Join(extractedDir, "artifacts/layout"))
	if err != nil {
		return errors.Wrapf(err, "failed to parse OCI Layout from archive %s", opts.ArchiveFile)
	}

	// Push updated images (renamed based on provided bundle tag) with same digests
	// then update the bundle with new values (image name, digest)
	for i, invImg := range bun.InvocationImages {
		newImgName, err := getNewImageNameFromBundleReference(invImg.Image, opts.Reference)
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
		newImgName, err := getNewImageNameFromBundleReference(img.Image, opts.Reference)
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

	bundleRef := cnab.BundleReference{
		Reference:  opts.ref,
		Definition: bun,
	}
	bundleRef, err = p.Registry.PushBundle(bundleRef, opts.InsecureRegistry)
	if err != nil {
		return err
	}

	// Perhaps we have a cached version of a bundle with the same tag, previously pulled
	// If so, replace it, as it is most likely out-of-date per this publish
	return p.refreshCachedBundle(bundleRef)
}

// extractBundle extracts a bundle using the provided opts and returnsthe extracted bundle
func (p *Porter) extractBundle(tmpDir, source string) (bundle.Bundle, error) {
	if p.Debug {
		fmt.Fprintf(p.Err, "Extracting bundle from archive %s...\n", source)
	}

	l := loader.NewLoader()
	imp := packager.NewImporter(source, tmpDir, l)
	err := imp.Import()
	if err != nil {
		return bundle.Bundle{}, errors.Wrapf(err, "failed to extract bundle from archive %s", source)
	}

	bun, err := l.Load(filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"), "bundle.json"))
	if err != nil {
		return bundle.Bundle{}, errors.Wrapf(err, "failed to load bundle from archive %s", source)
	}

	return *bun, nil
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
func (p *Porter) updateBundleWithNewImage(bun bundle.Bundle, newImg image.Name, digest image.Digest, index interface{}) error {
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

// getNewImageNameFromBundleReference derives a new image.Name object from the provided original
// image (string) using the provided bundleTag to glean registry/org/etc.
func getNewImageNameFromBundleReference(origImg, bundleTag string) (image.Name, error) {
	origName, err := image.NewName(origImg)
	if err != nil {
		return image.EmptyName, errors.Wrapf(err, "unable to parse image %q into domain/path components", origImg)
	}

	bundleName, err := image.NewName(bundleTag)
	if err != nil {
		return image.EmptyName, errors.Wrapf(err, "unable to parse bundle tag %q into domain/path components", bundleTag)
	}

	// Use the image name with the bundle location
	newImg := path.Join(path.Dir(bundleName.Name()), path.Base(origName.Name()))

	newImgName, err := image.NewName(newImg)
	if err != nil {
		return image.EmptyName, errors.Wrapf(err, "unable to parse image %q into domain/path components", newImg)
	}

	return newImgName, nil
}

func (p *Porter) rewriteBundleWithInvocationImageDigest(digest digest.Digest, manifestPath string) (bundle.Bundle, error) {
	taggedImage, err := p.rewriteImageWithDigest(p.Manifest.Image, digest.String())
	if err != nil {
		return bundle.Bundle{}, errors.Wrap(err, "unable to update invocation image reference")
	}

	fmt.Fprintln(p.Out, "\nRewriting CNAB bundle.json...")
	err = p.buildBundle(taggedImage, digest)
	if err != nil {
		return bundle.Bundle{}, errors.Wrap(err, "unable to rewrite CNAB bundle.json with updated invocation image digest")
	}

	b, err := p.FileSystem.ReadFile(build.LOCAL_BUNDLE)
	bun, err := bundle.ParseReader(bytes.NewBuffer(b))
	if err != nil {
		return bundle.Bundle{}, errors.Wrap(err, "unable to load CNAB bundle")
	}

	return bun, nil
}

func (p *Porter) rewriteImageWithDigest(InvocationImage string, imgDigest string) (string, error) {
	ref, err := cnab.ParseOCIReference(InvocationImage)
	if err != nil {
		return "", fmt.Errorf("unable to parse docker image: %s", err)
	}
	digestedRef, err := ref.WithDigest(digest.Digest(imgDigest))
	if err != nil {
		return "", err
	}
	return digestedRef.String(), nil
}

// refreshCachedBundle will store a bundle anew, if a bundle with the same tag is found in the cache
func (p *Porter) refreshCachedBundle(bundleRef cnab.BundleReference) error {
	if _, found, _ := p.Cache.FindBundle(bundleRef.Reference); found {
		_, err := p.Cache.StoreBundle(bundleRef)
		if err != nil {
			fmt.Fprintf(p.Err, "warning: unable to update cache for bundle %s: %s\n", bundleRef.Reference, err)
		}
	}
	return nil
}
