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
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/cnabio/cnab-go/packager"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/opencontainers/go-digest"
)

// PublishOptions are options that may be specified when publishing a bundle.
// Porter handles defaulting any missing values.
type PublishOptions struct {
	BundlePullOptions
	BundleDefinitionOptions
	Tag         string
	Registry    string
	ArchiveFile string
	SignBundle  bool
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
		err := o.BundleDefinitionOptions.Validate(cfg.Context)
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

// Publish is a composite function that publishes an bundle image, rewrites the porter manifest
// and then regenerates the bundle.json. Finally, it publishes the manifest to an OCI registry.
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

	buildOpts := BuildOptions{
		BundleDefinitionOptions: opts.BundleDefinitionOptions,
		InsecureRegistry:        opts.InsecureRegistry,
	}
	bundleRef, err := p.ensureLocalBundleIsUpToDate(ctx, buildOpts)
	if err != nil {
		return err
	}

	// If the manifest file is the default/user-supplied manifest,
	// hot-swap in Porter's canonical translation (if exists) from
	// the .cnab/app directory, as there may be dynamic overrides for
	// the name and version fields to inform bundle image naming.
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

	// Capture original bundle image name as it may be updated below
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

	// Update bundle image and reference with opts.Reference, which may be
	// empty, which is fine - we still may need to pick up tag and/or registry
	// overrides
	if err := m.SetBundleImageAndReference(opts.Reference); err != nil {
		return log.Errorf("unable to set bundle image name and reference: %w", err)
	}

	if origInvImg != m.Image {
		// Tag it so that it will be known/found by Docker for publishing
		builder := p.GetBuilder(ctx)
		if err := builder.TagBundleImage(ctx, origInvImg, m.Image); err != nil {
			return err
		}
	}

	if m.Reference == "" {
		return log.Errorf("porter.yaml is missing registry or reference values needed for publishing")
	}

	// var bundleRef cnab.BundleReference
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
		return log.Errorf("unable to push bundle image %q: %w", m.Image, err)
	}

	stamp, err := configadapter.LoadStamp(bundleRef.Definition)
	if err != nil {
		return log.Errorf("failed to load stamp from bundle definition: %w", err)
	}
	bundleRef.Definition, err = p.rewriteBundleWithBundleImageDigest(ctx, m, bundleRef.Digest, stamp.PreserveTags)
	if err != nil {
		return err
	}

	bundleRef, err = p.Registry.PushBundle(ctx, bundleRef, regOpts)
	if err != nil {
		return err
	}

	if opts.SignBundle {
		log.Debugf("signing bundle %s", bundleRef.String())
		inImage, err := cnab.CalculateTemporaryImageTag(bundleRef.Reference)
		if err != nil {
			return log.Errorf("error calculation temporary image tag: %w", err)
		}
		log.Debugf("Signing bundle image %s.", inImage.String())
		err = p.signImage(ctx, inImage)
		if err != nil {
			return log.Errorf("error signing bundle image: %w", err)
		}
		log.Debugf("Signing bundle artifact %s.", bundleRef.Reference.String())
		err = p.signImage(ctx, bundleRef.Reference)
		if err != nil {
			return log.Errorf("error signing bundle artifact: %w", err)
		}
	}

	// Perhaps we have a cached version of a bundle with the same reference, previously pulled
	// If so, replace it, as it is most likely out-of-date per this publish
	err = p.refreshCachedBundle(bundleRef)
	return log.Error(err)
}

// publishFromArchive (re-)publishes a bundle, provided by the archive file, using the provided tag.
//
// After the bundle is extracted from the archive, we iterate through all of the images (bundle
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
	defer func() {
		err = errors.Join(err, p.FileSystem.RemoveAll(tmpDir))
	}()

	bundleRef, err := p.extractBundle(ctx, tmpDir, source)
	if err != nil {
		return err
	}

	bundleRef.Reference = ref

	log.Infof("Beginning bundle publish to %s. This may take some time.", opts.Reference)

	// Read the extracted OCI Layout
	extractedDir := filepath.Join(tmpDir, strings.TrimSuffix(filepath.Base(source), ".tgz"))
	layoutPath, err := layout.FromPath(filepath.Join(extractedDir, "artifacts/layout"))
	if err != nil {
		return log.Errorf("failed to parse OCI Layout from archive %s: %w", opts.ArchiveFile, err)
	}

	// Push updated images (renamed based on provided bundle tag) with same digests
	// then update the bundle with new values (image name, digest)
	for _, invImg := range bundleRef.Definition.InvocationImages {
		relocMap, err := p.relocateImage(ctx, bundleRef.RelocationMap, layoutPath, invImg.Image, opts.Reference, regOpts)
		if err != nil {
			return log.Error(err)
		}

		bundleRef.RelocationMap = relocMap

		if opts.SignBundle {
			relocInvImage := relocMap[invImg.Image]
			log.Debugf("Signing bundle image %s...", relocInvImage)
			invImageRef, err := cnab.ParseOCIReference(relocInvImage)
			if err != nil {
				return log.Errorf("failed to parse OCI reference %s: %w", relocInvImage, err)
			}
			err = p.signImage(ctx, invImageRef)
			if err != nil {
				return log.Errorf("failed to sign image %s: %w", invImageRef.String(), err)
			}
		}
	}
	for _, img := range bundleRef.Definition.Images {
		relocMap, err := p.relocateImage(ctx, bundleRef.RelocationMap, layoutPath, img.Image, opts.Reference, regOpts)
		if err != nil {
			return log.Error(err)
		}

		bundleRef.RelocationMap = relocMap
	}

	bundleRef, err = p.Registry.PushBundle(ctx, bundleRef, regOpts)
	if err != nil {
		return err
	}

	if opts.SignBundle {
		log.Debugf("Signing bundle %s...", bundleRef.String())
		err = p.signImage(ctx, bundleRef.Reference)
		if err != nil {
			return log.Errorf("failed to sign bundle %s: %w", bundleRef.String(), err)
		}
	}

	// Perhaps we have a cached version of a bundle with the same tag, previously pulled
	// If so, replace it, as it is most likely out-of-date per this publish
	err = p.refreshCachedBundle(bundleRef)
	return log.Error(err)
}

// extractBundle extracts a bundle using the provided opts and returns the extracted bundle
func (p *Porter) extractBundle(ctx context.Context, tmpDir, source string) (cnab.BundleReference, error) {
	_, span := tracing.StartSpan(ctx)
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
func (p *Porter) pushUpdatedImage(ctx context.Context, layoutPath layout.Path, origImg string, destRef name.Reference, opts cnabtooci.RegistryOptions) (v1.Hash, error) {
	// Find image in layout
	digest, err := findImageInLayout(layoutPath, origImg)
	if err != nil {
		return v1.Hash{}, fmt.Errorf("unable to find image %s in archived OCI Layout: %w", origImg, err)
	}

	// Push to new location
	err = p.pushImageFromLayout(ctx, layoutPath, digest, destRef, opts)
	if err != nil {
		return v1.Hash{}, err
	}

	return digest, nil
}

// findImageInLayout searches for an image in the OCI layout by matching the image name.
// It returns the digest of the matching image.
func findImageInLayout(layoutPath layout.Path, imageName string) (v1.Hash, error) {
	// Get the image index from the layout
	index, err := layoutPath.ImageIndex()
	if err != nil {
		return v1.Hash{}, fmt.Errorf("unable to read image index from layout: %w", err)
	}

	// Get the index manifest
	indexManifest, err := index.IndexManifest()
	if err != nil {
		return v1.Hash{}, fmt.Errorf("unable to read index manifest: %w", err)
	}

	// Search through descriptors for a matching image
	// Try to match by annotation org.opencontainers.image.ref.name
	for _, desc := range indexManifest.Manifests {
		if desc.Annotations != nil {
			if annotName, ok := desc.Annotations["org.opencontainers.image.ref.name"]; ok {
				// Try exact match first
				if annotName == imageName {
					return desc.Digest, nil
				}

				// Try matching without registry (e.g., "myorg/myapp:v1.0" matches "registry.io/myorg/myapp:v1.0")
				annotRef, err := name.ParseReference(annotName, name.WeakValidation)
				if err == nil {
					// Compare repository and tag/digest parts
					searchRef, searchErr := name.ParseReference(imageName, name.WeakValidation)
					if searchErr == nil {
						// Check if the repository paths match (ignoring registry)
						annotCtx := annotRef.Context()
						searchCtx := searchRef.Context()
						if annotCtx.RepositoryStr() == searchCtx.RepositoryStr() {
							// Also check tag/digest if present
							annotID := annotRef.Identifier()
							searchID := searchRef.Identifier()
							if annotID == searchID {
								return desc.Digest, nil
							}
						}
					}
				}
			}
		}
	}

	return v1.Hash{}, fmt.Errorf("image %s not found in layout", imageName)
}

// pushImageFromLayout retrieves an image from the OCI layout by digest and pushes it to a destination registry
func (p *Porter) pushImageFromLayout(ctx context.Context, layoutPath layout.Path, digest v1.Hash, destRef name.Reference, opts cnabtooci.RegistryOptions) error {
	// Get image from layout by digest
	img, err := layoutPath.Image(digest)
	if err != nil {
		return fmt.Errorf("unable to get image from layout: %w", err)
	}

	// Build remote options (includes auth and insecure registry settings)
	remoteOpts := opts.ToRemoteOptions()

	// Push to destination registry
	err = remote.Write(destRef, img, remoteOpts...)
	if err != nil {
		return fmt.Errorf("unable to push image to %s: %w", destRef.String(), err)
	}

	return nil
}

// getNewImageNameFromBundleReference derives a new image reference from the provided original
// image (string) using the provided bundleTag to clean registry/org/etc.
func getNewImageNameFromBundleReference(origImg, bundleTag string) (name.Reference, error) {
	origImgRef, err := cnab.ParseOCIReference(origImg)
	if err != nil {
		return nil, err
	}

	bundleRef, err := cnab.ParseOCIReference(bundleTag)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Apply the temporary tag to the current bundle to determine the new location for the image
	newImgRef, err := bundleRef.WithTag(tmpImage.Tag())
	if err != nil {
		return nil, err
	}

	// Parse the reference string into a name.Reference
	return name.ParseReference(newImgRef.String())
}

func (p *Porter) rewriteBundleWithBundleImageDigest(ctx context.Context, m *manifest.Manifest, digest digest.Digest, preserveTags bool) (cnab.ExtendedBundle, error) {
	taggedImage, err := p.rewriteImageWithDigest(m.Image, digest.String())
	if err != nil {
		return cnab.ExtendedBundle{}, fmt.Errorf("unable to update bundle image reference: %w", err)
	}
	m.Image = taggedImage

	fmt.Fprintln(p.Out, "\nRewriting CNAB bundle.json...")
	err = p.buildBundle(ctx, m, digest, preserveTags)
	if err != nil {
		return cnab.ExtendedBundle{}, fmt.Errorf("unable to rewrite CNAB bundle.json with updated bundle image digest: %w", err)
	}

	bun, err := cnab.LoadBundle(p.Context, build.LOCAL_BUNDLE)
	if err != nil {
		return cnab.ExtendedBundle{}, fmt.Errorf("unable to load CNAB bundle: %w", err)
	}

	return bun, nil
}

func (p *Porter) relocateImage(ctx context.Context, relocationMap relocation.ImageRelocationMap, layoutPath layout.Path, originImg string, newReference string, opts cnabtooci.RegistryOptions) (relocation.ImageRelocationMap, error) {
	newImgRef, err := getNewImageNameFromBundleReference(originImg, newReference)
	if err != nil {
		return nil, err
	}

	originImgRef := originImg
	if relocatedImage, ok := relocationMap[originImg]; ok {
		originImgRef = relocatedImage
	}
	digest, err := p.pushUpdatedImage(ctx, layoutPath, originImgRef, newImgRef, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to push updated image: %w", err)
	}

	taggedImage, err := p.rewriteImageWithDigest(newImgRef.String(), digest.String())
	if err != nil {
		return nil, fmt.Errorf("unable to update image reference for %s: %w", newImgRef.String(), err)
	}

	// update relocation map
	relocationMap[originImg] = taggedImage
	return relocationMap, nil
}

func (p *Porter) rewriteImageWithDigest(image string, imgDigest string) (string, error) {
	taggedRef, err := cnab.ParseOCIReference(image)
	if err != nil {
		return "", fmt.Errorf("unable to parse docker image: %s", err)
	}

	// Change the bundle image from bundlerepo:tag-hash => bundlerepo@sha256:abc123
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

// signImage signs a image using the configured signing plugin
func (p *Porter) signImage(ctx context.Context, ref cnab.OCIReference) error {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Debugf("Signing image %s...", ref.String())
	return p.Signer.Sign(context.Background(), ref.String())
}
