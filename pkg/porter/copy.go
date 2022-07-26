package porter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

type CopyOpts struct {
	Source           string
	sourceRef        cnab.OCIReference
	Destination      string
	InsecureRegistry bool
}

// Validate performs validation logic on the options specified for a bundle copy
func (c *CopyOpts) Validate() error {
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
func (p *Porter) CopyBundle(ctx context.Context, c *CopyOpts) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("source", c.sourceRef.String()),
		attribute.String("destination", c.Destination),
	)
	defer span.EndSpan()

	destinationRef, err := generateNewBundleRef(c.sourceRef, c.Destination)
	if err != nil {
		return span.Error(err)
	}

	span.Infof("Beginning bundle copy to %s. This may take some time.", destinationRef)
	bunRef, err := p.Registry.PullBundle(ctx, c.sourceRef, c.InsecureRegistry)
	if err != nil {
		return span.Error(fmt.Errorf("unable to pull bundle before copying: %w", err))
	}

	bunRef.Reference = destinationRef
	_, err = p.Registry.PushBundle(context.Background(), bunRef, c.InsecureRegistry)
	if err != nil {
		return span.Error(fmt.Errorf("unable to copy bundle to new location: %w", err))
	}
	return nil
}
