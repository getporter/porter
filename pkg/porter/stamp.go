package porter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/tracing"
)

// ensureLocalBundleIsUpToDate ensures that the bundle is up to date with the porter manifest,
// if it is out-of-date, performs a build of the bundle.
func (p *Porter) ensureLocalBundleIsUpToDate(ctx context.Context, opts bundleFileOptions) (cnab.BundleReference, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	if opts.File == "" {
		return cnab.BundleReference{}, nil
	}

	upToDate, err := p.IsBundleUpToDate(ctx, opts)
	if err != nil {
		fmt.Fprintln(p.Err, "warning", err)
	}

	if !upToDate {
		fmt.Fprintln(p.Out, "Building bundle ===>")
		// opts.File is non-empty, which overrides opts.CNABFile if set
		// (which may be if a cached bundle is fetched e.g. when running an action)
		opts.CNABFile = ""
		buildOpts := BuildOptions{bundleFileOptions: opts}
		buildOpts.Validate(p)
		err := p.Build(ctx, buildOpts)
		if err != nil {
			return cnab.BundleReference{}, err
		}
	}

	bun, err := cnab.LoadBundle(p.Context, build.LOCAL_BUNDLE)
	if err != nil {
		return cnab.BundleReference{}, err
	}

	return cnab.BundleReference{
		Definition: bun,
	}, nil
}

// IsBundleUpToDate checks the hash of the manifest against the hash in cnab/bundle.json.
func (p *Porter) IsBundleUpToDate(ctx context.Context, opts bundleFileOptions) (bool, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if opts.File == "" {
		return false, span.Error(errors.New("File is required"))
	}
	m, err := manifest.LoadManifestFrom(ctx, p.Config, opts.File)
	if err != nil {
		return false, err
	}

	if exists, _ := p.FileSystem.Exists(opts.CNABFile); exists {
		bun, err := cnab.LoadBundle(p.Context, opts.CNABFile)
		if err != nil {
			return false, span.Error(fmt.Errorf("could not marshal data from %s: %w", opts.CNABFile, err))
		}

		// Check whether invocation images exist in host registry.
		for _, invocationImage := range bun.InvocationImages {
			// if the invovationImage is built before using a random string tag,
			// we should rebuild it with the new format
			if strings.HasSuffix(invocationImage.Image, "-installer") {
				return false, nil
			}

			imgRef, err := cnab.ParseOCIReference(invocationImage.Image)
			if err != nil {
				return false, span.Errorf("error parsing %s as an OCI image reference: %w", invocationImage.Image, err)
			}

			cachedImg, err := p.Registry.GetCachedImage(ctx, imgRef)
			if err != nil {
				return false, err
			}

			if cachedImg.IsZero() {
				span.Debugf("Invocation image %s doesn't exist in the local image cache, will need to build first", invocationImage.Image)
				return false, nil
			}
		}

		oldStamp, err := configadapter.LoadStamp(bun)
		if err != nil {
			return false, span.Error(fmt.Errorf("could not load stamp from %s: %w", opts.CNABFile, err))
		}

		mixins, err := p.getUsedMixins(ctx, m)
		if err != nil {
			return false, fmt.Errorf("error while listing used mixins: %w", err)
		}

		converter := configadapter.NewManifestConverter(p.Config, m, nil, mixins)
		newDigest, err := converter.DigestManifest()
		if err != nil {
			span.Debugf("could not determine if the bundle is up-to-date so will rebuild just in case: %w", err)
			return false, nil
		}
		return oldStamp.ManifestDigest == newDigest, nil
	}

	return false, nil
}
