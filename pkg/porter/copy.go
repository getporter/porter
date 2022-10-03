package porter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

type CopyOpts struct {
	Source           string
	sourceRef        cnab.OCIReference
	Destination      string
	InsecureRegistry bool
	Force            bool
}

// Validate performs validation logic on the options specified for a bundle copy
func (c *CopyOpts) Validate(cfg *config.Config) error {
	var err error
	if c.Destination == "" {
		return errors.New("--destination is required")
	}

	c.sourceRef, err = cnab.ParseOCIReference(c.Source)
	if err != nil {
		return fmt.Errorf("invalid value for --source, specified value should be of the form REGISTRY/bundle:tag or REGISTRY/bundle@sha: %w", err)
	}
	if c.sourceRef.HasDigest() && isCopyReferenceOnly(c.Destination) {
		return errors.New("--destination must be tagged reference when --source is digested reference")
	}

	// Apply the global config for force overwrite
	if !c.Force && cfg.Data.ForceOverwrite {
		c.Force = true
	}

	return nil
}

func isCopyReferenceOnly(dest string) bool {
	ref, err := cnab.ParseOCIReference(dest)
	if err != nil {
		return false
	}
	return ref.IsRepositoryOnly()
}

func generateNewBundleRef(source cnab.OCIReference, dest string) (cnab.OCIReference, error) {
	if isCopyReferenceOnly(dest) {
		srcVal := source.String()
		bundleNameRef := srcVal[strings.LastIndex(srcVal, "/")+1:]
		dest = fmt.Sprintf("%s/%s", dest, bundleNameRef)
	}
	return cnab.ParseOCIReference(dest)
}

// CopyBundle copies a bundle from one repository to another
func (p *Porter) CopyBundle(ctx context.Context, opts *CopyOpts) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("source", opts.sourceRef.String()),
		attribute.String("destination", opts.Destination),
	)
	defer span.EndSpan()

	destinationRef, err := generateNewBundleRef(opts.sourceRef, opts.Destination)
	if err != nil {
		return span.Error(err)
	}

	regOpts := cnabtooci.RegistryOptions{
		InsecureRegistry: opts.InsecureRegistry,
	}

	// Before we attempt to push, check if it already exists in the destination registry
	if !opts.Force {
		_, err := p.Registry.GetBundleMetadata(ctx, destinationRef, regOpts)
		if err != nil {
			if !errors.Is(err, cnabtooci.ErrNotFound{}) {
				return span.Errorf("Copy stopped because detection of %s in the destination registry failed. To overwrite it, repeat the command with --force specified: %w", destinationRef, err)
			}
		} else {
			return span.Errorf("Copy stopped because %s already exists in the destination registry. To overwrite it, repeat the command with --force specified.", destinationRef)
		}
	}

	span.Infof("Beginning bundle copy to %s. This may take some time.", destinationRef)
	bunRef, err := p.Registry.PullBundle(ctx, opts.sourceRef, regOpts)
	if err != nil {
		return span.Error(fmt.Errorf("unable to pull bundle before copying: %w", err))
	}

	bunRef.Reference = destinationRef

	_, err = p.Registry.PushBundle(ctx, bunRef, regOpts)
	if err != nil {
		return span.Error(fmt.Errorf("unable to copy bundle to new location: %w", err))
	}
	return nil
}
