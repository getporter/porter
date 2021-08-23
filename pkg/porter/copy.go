package porter

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "invalid value for --source, specified value should be of the form REGISTRY/bundle:tag or REGISTRY/bundle@sha")
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
func (p *Porter) CopyBundle(c *CopyOpts) error {
	destinationRef, err := generateNewBundleRef(c.sourceRef, c.Destination)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "Beginning bundle copy to %s. This may take some time.\n", destinationRef)
	bunRef, err := p.Registry.PullBundle(c.sourceRef, c.InsecureRegistry)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before copying")
	}

	bunRef.Reference = destinationRef
	_, err = p.Registry.PushBundle(bunRef, c.InsecureRegistry)
	if err != nil {
		return errors.Wrap(err, "unable to copy bundle to new location")
	}
	return nil
}
