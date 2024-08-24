package porter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// ensureLocalBundleIsUpToDate ensures that the bundle is up-to-date with the porter manifest,
// if it is out-of-date, performs a build of the bundle.
func (p *Porter) ensureLocalBundleIsUpToDate(ctx context.Context, opts BuildOptions) (cnab.BundleReference, error) {
	ctx, log := tracing.StartSpan(ctx,
		attribute.Bool("autobuild-disabled", opts.AutoBuildDisabled))
	defer log.EndSpan()

	if opts.File == "" {
		return cnab.BundleReference{}, nil
	}

	upToDate, err := p.IsBundleUpToDate(ctx, opts.BundleDefinitionOptions)
	if err != nil {
		log.Warnf("WARNING: %w", err)
	}

	if !upToDate {
		if opts.AutoBuildDisabled {
			log.Warn("WARNING: The bundle is out-of-date. Skipping autobuild because --autobuild-disabled was specified")
		} else {
			log.Info("Changes have been detected and the previously built bundle is out-of-date, rebuilding the bundle before proceeding...")
			log.Info("Building bundle ===>")
			// opts.File is non-empty, which overrides opts.CNABFile if set
			// (which may be if a cached bundle is fetched e.g. when running an action)
			opts.CNABFile = ""
			buildOpts := opts
			if err = buildOpts.Validate(p); err != nil {
				return cnab.BundleReference{}, log.Errorf("Validation of build options when autobuilding the bundle failed: %w", err)
			}
			err := p.Build(ctx, buildOpts)
			if err != nil {
				return cnab.BundleReference{}, err
			}
		}
	}

	bun, err := cnab.LoadBundle(p.Context, build.LOCAL_BUNDLE)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && opts.AutoBuildDisabled {
			return cnab.BundleReference{}, log.Errorf("Attempted to use a bundle from source without building it first when --autobuild-disabled is set. Build the bundle and try again: %w", err)
		}
		return cnab.BundleReference{}, log.Error(err)
	}

	return cnab.BundleReference{
		Definition: bun,
	}, nil
}

// IsBundleUpToDate checks the hash of the manifest against the hash in cnab/bundle.json.
func (p *Porter) IsBundleUpToDate(ctx context.Context, opts BundleDefinitionOptions) (bool, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debugf("Checking if the bundle is up-to-date...")

	// This is a prefix for any message that explains why the bundle is out-of-date
	const rebuildMessagePrefix = "Bundle is out-of-date and must be rebuilt"

	if opts.File == "" {
		span.Debugf("%s because the current bundle was not specified. Please report this as a bug!", rebuildMessagePrefix)
		return false, span.Errorf("File is required")
	}
	m, err := manifest.LoadManifestFrom(ctx, p.Config, opts.File)
	if err != nil {
		err = fmt.Errorf("the current bundle could not be read: %w", err)
		span.Debugf("%s: %w", rebuildMessagePrefix, err)
		return false, span.Error(err)
	}

	if exists, _ := p.FileSystem.Exists(opts.CNABFile); exists {
		bun, err := cnab.LoadBundle(p.Context, opts.CNABFile)
		if err != nil {
			err = fmt.Errorf("the previously built bundle at %s could not be read: %w", opts.CNABFile, err)
			span.Debugf("%s: %w", rebuildMessagePrefix, err)
			return false, span.Error(err)
		}

		// Check whether bundle images exist in host registry.
		for _, invocationImage := range bun.InvocationImages {
			// if the invocationImage is built before using a random string tag,
			// we should rebuild it with the new format
			if strings.HasSuffix(invocationImage.Image, "-installer") {
				span.Debugf("%s because it uses the old -installer suffixed image name (%s)", invocationImage.Image)
				return false, nil
			}

			imgRef, err := cnab.ParseOCIReference(invocationImage.Image)
			if err != nil {
				err = fmt.Errorf("error parsing %s as an OCI image reference: %w", invocationImage.Image, err)
				span.Debugf("%s: %w", rebuildMessagePrefix, err)
				return false, span.Error(err)
			}

			_, err = p.Registry.GetCachedImage(ctx, imgRef)
			if err != nil {
				if errors.Is(err, cnabtooci.ErrNotFound{}) {
					span.Debugf("%s because the bundle image %s doesn't exist in the local image cache", rebuildMessagePrefix, invocationImage.Image)
					return false, nil
				}
				err = fmt.Errorf("an error occurred checking the Docker cache for the bundle image: %w", err)
				span.Debugf("%s: %w", rebuildMessagePrefix, err)
				return false, span.Error(err)
			}
		}

		oldStamp, err := configadapter.LoadStamp(bun)
		if err != nil {
			err = fmt.Errorf("could not load stamp from %s: %w", opts.CNABFile, err)
			span.Debugf("%s: %w", rebuildMessagePrefix)
			return false, span.Error(err)
		}

		mixins, err := p.getUsedMixins(ctx, m)
		if err != nil {
			err = fmt.Errorf("an error occurred while listing used mixins: %w", err)
			span.Debugf("%s: %w", rebuildMessagePrefix, err)
			return false, span.Error(err)
		}

		converter := configadapter.NewManifestConverter(p.Config, m, nil, mixins)
		newDigest, err := converter.DigestManifest()
		if err != nil {
			err = fmt.Errorf("the current manifest digest cannot be calculated: %w", err)
			span.Debugf("%s: %w", rebuildMessagePrefix, err)
			return false, span.Error(err)
		}

		manifestChanged := oldStamp.ManifestDigest != newDigest
		if manifestChanged {
			span.Debugf("%s because the cached bundle is stale", rebuildMessagePrefix)
			if span.IsTracingEnabled() {
				previousStampB, _ := json.Marshal(oldStamp)
				currentStamp, _ := converter.GenerateStamp(ctx)
				currentStampB, _ := json.Marshal(currentStamp)
				span.SetAttributes(
					attribute.String("previous-stamp", string(previousStampB)),
					attribute.String("current-stamp", string(currentStampB)),
				)
			}
			return false, nil
		}

		span.Debugf("Bundle is up-to-date!")
		return true, nil
	}

	span.Debugf("%s because a previously built bundle was not found", rebuildMessagePrefix)
	return false, nil
}
