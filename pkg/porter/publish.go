package porter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/cnabio/cnab-go/packager"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/cnabio/image-relocation/pkg/image"
	"github.com/cnabio/image-relocation/pkg/registry"
	"github.com/cnabio/image-relocation/pkg/registry/ggcr"
	"github.com/opencontainers/go-digest"
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
func (o *PublishOptions) Validate(cfg *config.Config) error {
	if o.ArchiveFile != "" {
		// Verify the archive file can be accessed
		if _, err := cfg.FileSystem.Stat(o.ArchiveFile); err != nil {
			return fmt.Errorf("unable to access --archive %s: %w", o.ArchiveFile, err)
		}

		if o.Reference == "" {
			return errors.New("must provide a value for --reference of the form REGISTRY/bundle:tag")
		}
	} else {
		// Proceed with publishing from the resolved build context directory
		err := o.bundleFileOptions.Validate(cfg.Context)
		if err != nil {
			return err
		}

		if o.File == "" {
			return fmt.Errorf("could not find porter.yaml in the current directory %s, make sure you are in the right directory or specify the porter manifest with --file", o.Dir)
		}
	}

	if o.Reference != "" {
		return o.BundlePullOptions.Validate()
	}

	if o.Tag != "" {
		return o.validateTag()
	}

	// Apply the global config for force overwrite
	if !o.Force && cfg.Data.ForceOverwrite {
		o.Force = true
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
func (p *Porter) Publish(ctx context.Context, opts PublishOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	if opts.ArchiveFile == "" {
		return p.publishFromFile(ctx, opts)
	}
	return p.publishFromArchive(ctx, opts)
}

func (p *Porter) publishFromFile(ctx context.Context, opts PublishOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	_, err := p.ensureLocalBundleIsUpToDate(ctx, opts.bundleFileOptions)
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
		return log.Errorf("error reading manifest %s: %w", canonicalManifest)
	}

	var m *manifest.Manifest
	if canonicalExists {
		m, err = manifest.LoadManifestFrom(ctx, p.Config, canonicalManifest)
		if err != nil {
			return err
		}

		// We still want the user-provided manifest path to be tracked,
		// not Porter's canonical manifest path, for digest matching/auto-rebuilds
		m.ManifestPath = opts.File
	} else {
		m, err = manifest.LoadManifestFrom(ctx, p.Config, opts.File)
		if err != nil {
			return err
		}
	}

	// Capture original invocation image name as it may be updated below
	origInvImg := m.Image

	// Check for tag and registry overrides optionally supplied on publish
	if opts.Tag != "" {
		m.DockerTag = opts.Tag
	}
	if opts.Registry != "" {
		m.Registry = opts.Registry
	}

	// If either are non-empty, null out the reference on the manifest, as
	// it needs to be rebuilt with new values
	if opts.Tag != "" || opts.Registry != "" {
		m.Reference = ""
	}

	// Update invocation image and reference with opts.Reference, which may be
	// empty, which is fine - we still may need to pick up tag and/or registry
	// overrides
	if err := m.SetInvocationImageAndReference(opts.Reference); err != nil {
		return log.Errorf("unable to set invocation image name and reference: %w", err)
	}

	if origInvImg != m.Image {
		// Tag it so that it will be known/found by Docker for publishing
		builder := p.GetBuilder(ctx)
		if err := builder.TagInvocationImage(ctx, origInvImg, m.Image); err != nil {
			return err
		}
	}

	if m.Reference == "" {
		return log.Errorf("porter.yaml is missing registry or reference values needed for publishing")
	}

	var bundleRef cnab.BundleReference
	bundleRef.Reference, err = cnab.ParseOCIReference(m.Reference)
	if err != nil {
		return log.Errorf("invalid reference %s: %w", m.Reference, err)
	}

	imgRef, err := cnab.ParseOCIReference(m.Image)
	if err != nil {
		return log.Errorf("error parsing %s as an OCI reference: %w", m.Image, err)
	}

	regOpts := cnabtooci.RegistryOptions{
		InsecureRegistry: opts.InsecureRegistry,
	}

	// Before we attempt to push, check if any of the bundle exists already.
	// If force was not specified, we shouldn't push any of the bundle since
	// the bundle and images must be pushed as a unit.
	if !opts.Force {
		_, err := p.Registry.GetBundleMetadata(ctx, bundleRef.Reference, regOpts)
		if err != nil {
			if !errors.Is(err, cnabtooci.ErrNotFound{}) {
				return log.Errorf("Publish stopped because detection of %s in the destination registry failed. To overwrite it, repeat the command with --force specified: %w", bundleRef, err)
			}
		} else {
			return log.Errorf("Publish stopped because %s already exists in the destination registry. To overwrite it, repeat the command with --force specified.", bundleRef)
		}
	}

	bundleRef.Digest, err = p.Registry.PushImage(ctx, imgRef, regOpts)
	if err != nil {
		return log.Errorf("unable to push CNAB invocation image %q: %w", m.Image, err)
	}

	bundleRef.Definition, err = p.rewriteBundleWithInvocationImageDigest(ctx, m, bundleRef.Digest)
	if err != nil {
		return err
	}

	bundleRef, err = p.Registry.PushBundle(ctx, bundleRef, regOpts)
	if err != nil {
		return err
	}

	// Perhaps we have a cached version of a bundle with the same reference, previously pulled
	// If so, replace it, as it is most likely out-of-date per this publish
	err = p.refreshCachedBundle(bundleRef)
	return log.Error(err)
}

// publishFromArchive (re-)publishes a bundle, provided by the archive file, using the provided tag.
//
// After the bundle is extracted from the archive, we iterate through all of the images (invocation
// and application) listed in the bundle, grab their digests by parsing the extracted
// OCI Layout, rename each based on the registry/org values derived from the provided tag
// and then push each updated image with the original digests
//
// Finally, we update the relocation map in the original bundle, based
// on the newly copied images, and then push the bundle using the provided tag.
// (Currently we use the docker/cnab-to-oci library for this logic.)
func (p *Porter) publishFromArchive(ctx context.Context, opts PublishOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	regOpts := cnabtooci.RegistryOptions{InsecureRegistry: opts.InsecureRegistry}

	// Before we attempt to push, check if any of the bundle exists already.
	// If force was not specified, we shouldn't push any of the bundle since
	// the bundle and images must be pushed as a unit.
	ref := opts.GetReference()
	if !opts.Force {
		_, err := p.Registry.GetBundleMetadata(ctx, ref, regOpts)
		if err != nil {
			if !errors.Is(err, cnabtooci.ErrNotFound{}) {
				return log.Errorf("Publish stopped because detection of %s in the destination registry failed. To overwrite it, repeat the command with --force specified: %w", ref, err)
			}
		} else {
			return log.Errorf("Publish stopped because %s already exists in the destination registry. To overwrite it, repeat the command with --force specified.", ref)
		}
	}

	source := p.FileSystem.Abs(opts.ArchiveFile)
	tmpDir, err := p.FileSystem.TempDir("", "porter")
	if err != nil {
		return log.Errorf("error creating temp directory for archive extraction: %w", err)
	}
	defer p.FileSystem.RemoveAll(tmpDir)

	bundleRef, err := p.extractBundle(ctx, tmpDir, source)
	if err != nil {
		return err
	}

	bundleRef.Reference = ref

	log.Infof("Beginning bundle publish to %s. This may take some time.", opts.Reference)

	// Use the ggcr client to read the extracted OCI Layout
	extractedDir := filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"))
	var clientOpts []ggcr.Option
	if opts.InsecureRegistry {
		skipTLS := cnabtooci.GetInsecureRegistryTransport()
		clientOpts = append(clientOpts, ggcr.WithTransport(skipTLS))
	}
	client := ggcr.NewRegistryClient(clientOpts...)
	layout, err := client.ReadLayout(filepath.Join(extractedDir, "artifacts/layout"))
	if err != nil {
		return log.Errorf("failed to parse OCI Layout from archive %s: %w", opts.ArchiveFile, err)
	}

	// Push updated images (renamed based on provided bundle tag) with same digests
	// then update the bundle with new values (image name, digest)
	for _, invImg := range bundleRef.Definition.InvocationImages {
		relocMap, err := p.relocateImage(bundleRef.RelocationMap, layout, invImg.Image, opts.Reference)
		if err != nil {
			return log.Error(err)
		}

		bundleRef.RelocationMap = relocMap
	}
	for _, img := range bundleRef.Definition.Images {
		relocMap, err := p.relocateImage(bundleRef.RelocationMap, layout, img.Image, opts.Reference)
		if err != nil {
			return log.Error(err)
		}

		bundleRef.RelocationMap = relocMap
	}

	bundleRef, err = p.Registry.PushBundle(ctx, bundleRef, regOpts)
	if err != nil {
		return err
	}

	// Perhaps we have a cached version of a bundle with the same tag, previously pulled
	// If so, replace it, as it is most likely out-of-date per this publish
	err = p.refreshCachedBundle(bundleRef)
	return log.Error(err)
}

// extractBundle extracts a bundle using the provided opts and returns the extracted bundle
func (p *Porter) extractBundle(ctx context.Context, tmpDir, source string) (cnab.BundleReference, error) {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debugf("Extracting bundle from archive %s...", source)
	l := loader.NewLoader()
	imp := packager.NewImporter(source, tmpDir, l)
	err := imp.Import()
	if err != nil {
		return cnab.BundleReference{}, span.Error(fmt.Errorf("failed to extract bundle from archive %s: %w", source, err))
	}

	bun, err := l.Load(filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"), "bundle.json"))
	if err != nil {
		return cnab.BundleReference{}, span.Error(fmt.Errorf("failed to load bundle from archive %s: %w", source, err))
	}
	data, err := p.FileSystem.ReadFile(filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"), "relocation-mapping.json"))
	if err != nil {
		return cnab.BundleReference{}, span.Error(fmt.Errorf("failed to load relocation-mapping.json from archive %s: %w", source, err))
	}
	var reloMap relocation.ImageRelocationMap
	err = json.Unmarshal(data, &reloMap)
	if err != nil {
		return cnab.BundleReference{}, span.Error(fmt.Errorf("failed to parse relocation-mapping.json from archive %s: %w", source, err))
	}

	return cnab.BundleReference{Definition: cnab.ExtendedBundle{Bundle: *bun}, RelocationMap: reloMap}, nil
}

// pushUpdatedImage uses the provided layout to find the provided origImg,
// gathers the pre-existing digest and then pushes this digest using the newImgName
func pushUpdatedImage(layout registry.Layout, origImg string, newImgName image.Name) (image.Digest, error) {
	origImgName, err := image.NewName(origImg)
	if err != nil {
		return image.EmptyDigest, fmt.Errorf("unable to parse image %q into domain/path components: %w", origImg, err)
	}

	digest, err := layout.Find(origImgName)
	if err != nil {
		return image.EmptyDigest, fmt.Errorf("unable to find image %s in archived OCI Layout: %w", origImgName.String(), err)
	}

	err = layout.Push(digest, newImgName)
	if err != nil {
		return image.EmptyDigest, fmt.Errorf("unable to push image %s: %w", newImgName.String(), err)
	}

	return digest, nil
}

// getNewImageNameFromBundleReference derives a new image.Name object from the provided original
// image (string) using the provided bundleTag to clean registry/org/etc.
func getNewImageNameFromBundleReference(origImg, bundleTag string) (image.Name, error) {
	origImgRef, err := cnab.ParseOCIReference(origImg)
	if err != nil {
		return image.EmptyName, err
	}

	bundleRef, err := cnab.ParseOCIReference(bundleTag)
	if err != nil {
		return image.EmptyName, err
	}

	// Calculate a unique tag based on the original referenced image. It is safe to
	// use only the original image, and not a combination of both the destination and
	// the source to create a unique value, because we rewrite the referenced image
	// to always use a repository digest. The only time two images will have the same
	// source value is when they are the same image and have the same content. In
	// which case it is okay if two bundles both reference the same image and reuse
	// the same temporary tag because the content is the same.
	tmpImage, err := cnab.CalculateTemporaryImageTag(origImgRef)
	if err != nil {
		return image.EmptyName, err
	}

	// Apply the temporary tag to the current bundle to determine the new location for the image
	newImgRef, err := bundleRef.WithTag(tmpImage.Tag())
	if err != nil {
		return image.EmptyName, err
	}

	// Convert it to the relocation library's representation of an image reference
	return image.NewName(newImgRef.String())
}

func (p *Porter) rewriteBundleWithInvocationImageDigest(ctx context.Context, m *manifest.Manifest, digest digest.Digest) (cnab.ExtendedBundle, error) {
	taggedImage, err := p.rewriteImageWithDigest(m.Image, digest.String())
	if err != nil {
		return cnab.ExtendedBundle{}, fmt.Errorf("unable to update invocation image reference: %w", err)
	}
	m.Image = taggedImage

	fmt.Fprintln(p.Out, "\nRewriting CNAB bundle.json...")
	err = p.buildBundle(ctx, m, digest)
	if err != nil {
		return cnab.ExtendedBundle{}, fmt.Errorf("unable to rewrite CNAB bundle.json with updated invocation image digest: %w", err)
	}

	bun, err := cnab.LoadBundle(p.Context, build.LOCAL_BUNDLE)
	if err != nil {
		return cnab.ExtendedBundle{}, fmt.Errorf("unable to load CNAB bundle: %w", err)
	}

	return bun, nil
}

func (p *Porter) relocateImage(relocationMap relocation.ImageRelocationMap, layout registry.Layout, originImg string, newReference string) (relocation.ImageRelocationMap, error) {
	newImgName, err := getNewImageNameFromBundleReference(originImg, newReference)
	if err != nil {
		return nil, err
	}

	originImgRef := originImg
	if relocatedImage, ok := relocationMap[originImg]; ok {
		originImgRef = relocatedImage
	}
	digest, err := pushUpdatedImage(layout, originImgRef, newImgName)
	if err != nil {
		return nil, fmt.Errorf("unable to push updated image: %w", err)
	}

	taggedImage, err := p.rewriteImageWithDigest(newImgName.String(), digest.String())
	if err != nil {
		return nil, fmt.Errorf("unable to update image reference for %s: %w", newImgName.String(), err)
	}

	// update relocation map
	relocationMap[originImg] = taggedImage
	return relocationMap, nil
}

func (p *Porter) rewriteImageWithDigest(InvocationImage string, imgDigest string) (string, error) {
	taggedRef, err := cnab.ParseOCIReference(InvocationImage)
	if err != nil {
		return "", fmt.Errorf("unable to parse docker image: %s", err)
	}

	// Change the invocation image from bundlerepo:tag-hash => bundlerepo@sha256:abc123
	// Do not continue to reference the temporary tag that we used to push, otherwise that will prevent the registry from garbage collecting it later.
	repo := cnab.MustParseOCIReference(taggedRef.Repository())

	digestedRef, err := repo.WithDigest(digest.Digest(imgDigest))
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
