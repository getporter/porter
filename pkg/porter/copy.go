package porter

import (
	"fmt"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
)

type CopyOpts struct {
	Source           string
	Destination      string
	InsecureRegistry bool
}

// Validate performs validation logic on the options specified for a bundle copy
func (c *CopyOpts) Validate() error {
	if c.Destination == "" {
		return errors.New("--destination is required")
	}
	source, err := reference.ParseNormalizedNamed(c.Source)
	if err != nil {
		return errors.Wrap(err, "invalid value for --source, specified value should be of the form REGISTRY/bundle:tag or REGISTRY/bundle@sha")
	}
	if isCopyDigestReference(source) && isCopyReferenceOnly(c.Destination) {
		return errors.New("--destination must be tagged reference when --source is digested reference")
	}
	return nil
}

func isCopyDigestReference(source reference.Named) bool {
	if _, ok := source.(reference.Canonical); ok {
		return true
	}
	return false
}

func isCopyReferenceOnly(dest string) bool {
	named, err := reference.ParseNormalizedNamed(dest)
	if err != nil {
		return false
	}
	return reference.IsNameOnly(named)
}

func generateNewBundleRef(source, dest string) string {
	if isCopyReferenceOnly(dest) {
		bundleNameRef := source[strings.LastIndex(source, "/")+1:]
		return fmt.Sprintf("%s/%s", dest, bundleNameRef)
	}
	return dest
}

// CopyBundle copies a bundle from one repository to another
func (p *Porter) CopyBundle(c *CopyOpts) error {
	destinationRef := generateNewBundleRef(c.Source, c.Destination)
	fmt.Fprintf(p.Out, "Beginning bundle copy to %s. This may take some time.\n", destinationRef)
	bun, reloMap, err := p.Registry.PullBundle(c.Source, c.InsecureRegistry)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before copying")
	}
	_, err = p.Registry.PushBundle(bun, destinationRef, *reloMap, c.InsecureRegistry)
	if err != nil {
		return errors.Wrap(err, "unable to copy bundle to new location")
	}
	return nil
}
