package porter

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/experimental"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/docker/distribution/reference"
	"github.com/mikefarah/yq/v3/pkg/yqlib"
	"github.com/opencontainers/go-digest"
	"go.opentelemetry.io/otel/attribute"
)

// metadataOpts contain manifest fields eligible for dynamic
// updating prior to saving Porter's internal version of the manifest
type metadataOpts struct {
	Name    string
	Version string
}

// generateInternalManifest decodes the manifest designated by filepath and applies
// the provided generateInternalManifestOpts, saving the updated manifest to the path
// designated by build.LOCAL_MANIFEST
// if a referenced image does not have digest specified, update the manifest to use digest instead.
func (p *Porter) generateInternalManifest(ctx context.Context, opts BuildOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// Create the local app dir if it does not already exist
	err := p.FileSystem.MkdirAll(build.LOCAL_APP, pkg.FileModeDirectory)
	if err != nil {
		return span.Error(fmt.Errorf("unable to create directory %s: %w", build.LOCAL_APP, err))
	}

	e := yaml.NewEditor(p.FileSystem)
	err = e.ReadFile(opts.File)
	if err != nil {
		return span.Error(fmt.Errorf("unable to read manifest file %s: %w", opts.File, err))
	}

	if opts.Name != "" {
		if err = e.SetValue("name", opts.Name); err != nil {
			return err
		}
	}

	if opts.Version != "" {
		if err = e.SetValue("version", opts.Version); err != nil {
			return err
		}
	}

	for k, v := range opts.parsedCustoms {
		if err = e.SetValue("custom."+k, v); err != nil {
			return err
		}
	}

	regOpts := cnabtooci.RegistryOptions{
		InsecureRegistry: opts.InsecureRegistry,
	}

	// find all referenced images that does not have digest specified
	// get the image digest for all of them and update the manifest with the digest
	err = e.WalkNodes(ctx, "images.*", func(ctx context.Context, nc *yqlib.NodeContext) error {
		ctx, span := tracing.StartSpanWithName(ctx, "updateReferencedImageTagToDigest")
		defer span.EndSpan()

		img := &manifest.MappedImage{}
		if err := nc.Node.Decode(img); err != nil {
			return span.Errorf("failed to deserialize referenced image in manifest: %w", err)
		}

		span.SetAttributes(attribute.String("image", img.Repository))

		// if image digest is specified in the manifest, we don't need to get it
		// from registries
		if img.Digest != "" {
			return nil
		}

		ref, err := img.ToOCIReference()
		if err != nil {
			return span.Errorf("failed to parse image %s reference: %w", img.Repository, err)
		}
		if opts.PreserveTags {
			if img.Tag == "" {
				var path string
				for _, p := range nc.PathStack {
					switch t := p.(type) {
					case string:
						path += fmt.Sprintf("%s.", t)
					case int:
						path = strings.TrimSuffix(path, ".")
						path += fmt.Sprintf("[%s].", strconv.Itoa(t))
					default:
						continue
					}
				}

				return e.SetValue(path+"tag", "latest")
			}
		} else {

			digest, err := p.getImageDigest(ctx, ref, regOpts)
			if err != nil {
				return span.Error(err)
			}
			span.SetAttributes(attribute.String("digest", digest.Encoded()))

			var path string
			for _, p := range nc.PathStack {
				switch t := p.(type) {
				case string:
					path += fmt.Sprintf("%s.", t)
				case int:
					path = strings.TrimSuffix(path, ".")
					path += fmt.Sprintf("[%s].", strconv.Itoa(t))
				default:
					continue
				}
			}

			return e.SetValue(path+"digest", digest.String())
		}

		return nil
	})
	if err != nil {
		return err
	}

	if p.IsFeatureEnabled(experimental.FlagDependenciesV2) {
		if err = p.resolveDependencyDigest(ctx, e, regOpts); err != nil {
			return err
		}
	}

	return e.WriteFile(build.LOCAL_MANIFEST)
}

func (p *Porter) resolveDependencyDigest(ctx context.Context, e *yaml.Editor, opts cnabtooci.RegistryOptions) error {
	// find all referenced dependencies that does not have digest specified
	// get the digest for all of them and update the manifest with the digest
	return e.WalkNodes(ctx, "dependencies.requires.*", func(ctx context.Context, nc *yqlib.NodeContext) error {
		ctx, span := tracing.StartSpanWithName(ctx, "updateDependencyTagToDigest")
		defer span.EndSpan()

		dep := &manifest.Dependency{}
		if err := nc.Node.Decode(dep); err != nil {
			return span.Errorf("failed to deserialize dependency in manifest: %w", err)
		}

		span.SetAttributes(attribute.String("dependency", dep.Name))

		bundleOpts := BundleReferenceOptions{
			BundlePullOptions: BundlePullOptions{
				Reference:        dep.Bundle.Reference,
				InsecureRegistry: opts.InsecureRegistry,
			},
		}

		ref, err := cnab.ParseOCIReference(dep.Bundle.Reference)
		if err != nil {
			return span.Errorf("failed to parse OCI reference for dependency %s: %w", dep.Name, err)
		}

		if ref.Tag() == "" || ref.Tag() == "latest" {
			return nil
		}

		bundleRef, err := p.resolveBundleReference(ctx, &bundleOpts)
		if err != nil {
			return span.Errorf("failed to resolve dependency %s: %w", dep.Name, err)
		}

		digest := bundleRef.Digest
		span.SetAttributes(attribute.String("digest", digest.Encoded()))

		var path string
		for _, p := range nc.PathStack {
			switch t := p.(type) {
			case string:
				path += fmt.Sprintf("%s.", t)
			case int:
				path = strings.TrimSuffix(path, ".")
				path += fmt.Sprintf("[%s].", strconv.Itoa(t))
			default:
				continue
			}
		}

		newRef := cnab.OCIReference{
			Named: reference.TrimNamed(bundleRef.Reference.Named),
		}
		refWithDigest, err := newRef.WithDigest(digest)
		if err != nil {
			return span.Errorf("failed to set digest: %w", err)
		}

		return e.SetValue(path+"bundle.reference", refWithDigest.String())
	})
}

// getImageDigest retrieves the repository digest associated with the specified image reference.
func (p *Porter) getImageDigest(ctx context.Context, img cnab.OCIReference, regOpts cnabtooci.RegistryOptions) (digest.Digest, error) {
	ctx, span := tracing.StartSpan(ctx, attribute.String("image", img.String()))
	defer span.EndSpan()

	// if no image tag is specified, default to use latest
	if img.Tag() == "" {
		refWithTag, err := img.WithTag("latest")
		if err != nil {
			return "", span.Errorf("failed to create image reference %s with tag latest: %w", img.String(), err)
		}
		img = refWithTag
	}

	imgSummary, err := p.Registry.GetImageMetadata(ctx, img, regOpts)
	if err != nil {
		return "", err
	}

	imgDigest, err := imgSummary.GetRepositoryDigest()
	if err != nil {
		return "", span.Error(err)
	}

	span.SetAttributes(attribute.String("digest", imgDigest.String()))
	return imgDigest, nil
}
