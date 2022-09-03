package cnabtooci

import (
	"get.porter.sh/porter/pkg/cnab"
)

// BundleMetadata represents summary information about a bundle in a registry.
type BundleMetadata struct {
	// Reference to the bundle in a remote registry.
	Reference cnab.OCIReference

	// Digest is the
	Digest string
}
