package porter

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/pkg/yaml"
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

	e := yaml.NewEditor(p.Context)
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

		digest, err := p.getImageLatestDigest(ctx, ref)
		if err != nil {
			return span.Error(err)
		}

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

	})
	if err != nil {
		return err
	}

	return e.WriteFile(build.LOCAL_MANIFEST)
}

func (p *Porter) getImageLatestDigest(ctx context.Context, img cnab.OCIReference) (digest.Digest, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// if no image tag is specified, defautl to use latest
	if img.Tag() == "" {
		refWithTag, err := img.WithTag("latest")
		if err != nil {
			return "", span.Errorf("failed to create image reference %s with tag latest: %w", img.String(), err)
		}
		img = refWithTag
	}

	// Right now there isn't a way to specify --insecure-registry for build
	// because the underlying implementation in PullImage doesn't support it.
	err := p.Registry.PullImage(ctx, img, cnabtooci.RegistryOptions{})
	if err != nil {
		return "", err
	}

	imgSummary, err := p.Registry.GetCachedImage(ctx, img)
	if err != nil {
		return "", err
	}

	return imgSummary.Digest()
}
